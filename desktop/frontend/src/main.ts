import './style.css';
import './app.css';
import {ConnectAndLaunch, ConnectVPN, DisconnectVPN, GetStatus, SaveSettings} from '../wailsjs/go/main/App';

type Settings = {
    deviceLabel: string;
    targetAddress: string;
    tailscalePath: string;
    anyDeskPath: string;
};

type Status = {
    configured: boolean;
    vpnState: string;
    targetState: string;
    anyDeskState: string;
    message: string;
    updatedAt: string;
    settings: Settings;
};

const app = document.querySelector<HTMLDivElement>('#app')!;

app.innerHTML = `
  <main class="shell">
    <header class="topbar">
      <div class="brand">
        <div class="brand-mark" aria-hidden="true"><span></span><span></span></div>
        <div><h1>RemoteBridge</h1><p>Windows remote workspace</p></div>
      </div>
      <button class="icon-button" id="settingsButton" aria-label="연결 설정 열기">⚙</button>
    </header>

    <section class="hero">
      <div>
        <span class="eyebrow">SECURE REMOTE ACCESS</span>
        <h2 id="deviceTitle">원격 Windows PC</h2>
        <p id="deviceAddress">Tailscale 원격 장치를 설정해 주세요.</p>
      </div>
      <div class="overall" id="overall"><span class="pulse"></span><span>확인 중</span></div>
    </section>

    <section class="status-grid" aria-label="연결 상태">
      <article class="status-card">
        <div class="card-icon vpn">T</div>
        <div><span class="card-label">PRIVATE NETWORK</span><h3>Tailscale</h3><p>두 Windows PC의 사설 연결</p></div>
        <span class="badge" id="vpnBadge">확인 중</span>
      </article>
      <article class="status-card">
        <div class="card-icon target">PC</div>
        <div><span class="card-label">REMOTE DEVICE</span><h3>Windows 11</h3><p>AnyDesk TCP 7070</p></div>
        <span class="badge" id="targetBadge">확인 전</span>
      </article>
      <article class="status-card">
        <div class="card-icon anydesk">A</div>
        <div><span class="card-label">SCREEN CONTROL</span><h3>AnyDesk</h3><p>앱 자체 인증 사용</p></div>
        <span class="badge" id="anyDeskBadge">확인 중</span>
      </article>
    </section>

    <section class="action-panel">
      <div class="message-wrap">
        <span class="message-dot"></span>
        <div><strong id="message">상태를 불러오고 있습니다.</strong><small id="updatedAt"></small></div>
      </div>
      <div class="actions">
        <button class="button ghost" id="disconnectButton">VPN 종료</button>
        <button class="button secondary" id="vpnButton">VPN만 연결</button>
        <button class="button primary" id="connectButton"><span>연결하고 AnyDesk 열기</span><b>→</b></button>
      </div>
    </section>

    <footer><button class="text-button" id="refreshButton">↻ 상태 새로고침</button><span>RemoteBridge v3.0.0-alpha.2 · Tailscale 인증과 AnyDesk 암호는 앱에 저장하지 않습니다.</span></footer>
  </main>

  <dialog id="settingsDialog">
    <form id="settingsForm">
      <div class="dialog-head"><div><span class="eyebrow">CONNECTION PROFILE</span><h2>연결 설정</h2></div><button type="button" class="icon-button" id="closeSettings" aria-label="닫기">×</button></div>
      <p class="dialog-copy">두 PC를 같은 Tailnet에 로그인한 뒤 원격 장치의 Tailscale IP 또는 MagicDNS 이름을 입력하세요.</p>
      <div class="form-grid">
        <label><span>장치 이름</span><input name="deviceLabel" required maxlength="80" placeholder="Academy PC"></label>
        <label><span>원격 Tailscale 주소</span><input name="targetAddress" required placeholder="academy-pc 또는 100.x.x.x"></label>
        <label class="wide"><span>Tailscale 실행 파일 <em>선택</em></span><input name="tailscalePath" placeholder="자동 검색"></label>
        <label class="wide"><span>AnyDesk 실행 파일 <em>선택</em></span><input name="anyDeskPath" placeholder="자동 검색"></label>
      </div>
      <div class="dialog-actions"><button type="button" class="button ghost" id="cancelSettings">취소</button><button type="submit" class="button primary">설정 저장</button></div>
    </form>
  </dialog>
`;

const byId = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const dialog = byId<HTMLDialogElement>('settingsDialog');
const form = byId<HTMLFormElement>('settingsForm');
let current: Status | null = null;
let busy = false;

const labels: Record<string, [string, string]> = {
    running: ['연결됨', 'good'], stopped: ['꺼짐', 'muted'], starting: ['연결 중', 'warn'], 'needs-login': ['로그인 필요', 'warn'],
    reachable: ['응답함', 'good'], unreachable: ['응답 없음', 'bad'], unchecked: ['확인 전', 'muted'],
    ready: ['준비됨', 'good'], missing: ['찾지 못함', 'bad'], unknown: ['확인 중', 'muted'], error: ['오류', 'bad']
};

function setBadge(id: string, state: string) {
    const badge = byId<HTMLSpanElement>(id);
    const [text, tone] = labels[state] ?? [state, 'muted'];
    badge.textContent = text;
    badge.className = `badge ${tone}`;
}

function render(status: Status) {
    current = status;
    setBadge('vpnBadge', status.vpnState);
    setBadge('targetBadge', status.targetState);
    setBadge('anyDeskBadge', status.anyDeskState);
    byId('deviceTitle').textContent = status.settings?.deviceLabel || '원격 Windows PC';
    byId('deviceAddress').textContent = status.settings?.targetAddress ? `${status.settings.targetAddress} · Tailscale private network` : 'Tailscale 원격 장치를 설정해 주세요.';
    byId('message').textContent = status.message || '상태 확인을 완료했습니다.';
    byId('updatedAt').textContent = status.updatedAt ? `마지막 확인 ${new Date(status.updatedAt).toLocaleTimeString('ko-KR', {hour: '2-digit', minute: '2-digit', second: '2-digit'})}` : '';

    const online = status.vpnState === 'running' && status.targetState === 'reachable';
    const overall = byId('overall');
    overall.className = `overall ${online ? 'online' : ''}`;
    overall.innerHTML = `<span class="pulse"></span><span>${online ? '접속 가능' : '대기 중'}</span>`;
    byId<HTMLButtonElement>('disconnectButton').disabled = busy || status.vpnState !== 'running';
    byId<HTMLButtonElement>('connectButton').disabled = busy || !status.configured;
    byId<HTMLButtonElement>('vpnButton').disabled = busy || !status.configured || status.vpnState === 'running';
    if (!status.configured && !dialog.open) openSettings();
}

function setBusy(value: boolean) {
    busy = value;
    document.body.classList.toggle('busy', value);
    if (current) render(current);
}

async function run(operation: () => Promise<Status>) {
    if (busy) return;
    setBusy(true);
    try { render(await operation()); }
    catch { byId('message').textContent = '요청을 처리하지 못했습니다. 잠시 후 다시 시도하세요.'; }
    finally { setBusy(false); }
}

function openSettings() {
    const settings = current?.settings ?? {} as Settings;
    for (const key of ['deviceLabel', 'targetAddress', 'tailscalePath', 'anyDeskPath'] as const) {
        const input = form.elements.namedItem(key) as HTMLInputElement;
        input.value = settings[key] ?? '';
    }
    dialog.showModal();
}

byId('settingsButton').addEventListener('click', openSettings);
byId('closeSettings').addEventListener('click', () => dialog.close());
byId('cancelSettings').addEventListener('click', () => dialog.close());
byId('refreshButton').addEventListener('click', () => run(GetStatus));
byId('vpnButton').addEventListener('click', () => run(ConnectVPN));
byId('disconnectButton').addEventListener('click', () => run(DisconnectVPN));
byId('connectButton').addEventListener('click', () => run(ConnectAndLaunch));
form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const field = (name: string) => (form.elements.namedItem(name) as HTMLInputElement).value;
    const values: Settings = {
        deviceLabel: field('deviceLabel'),
        targetAddress: field('targetAddress'),
        tailscalePath: field('tailscalePath'),
        anyDeskPath: field('anyDeskPath')
    };
    setBusy(true);
    try {
        const status = await SaveSettings(values);
        render(status);
        if (status.configured) dialog.close();
    } finally { setBusy(false); }
});

GetStatus().then(render).catch(() => { byId('message').textContent = '앱 상태를 불러오지 못했습니다.'; });
