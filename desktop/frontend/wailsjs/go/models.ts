export namespace bridge {

	export class SettingsInput {
	    deviceLabel: string;
	    targetAddress: string;
	    tailscalePath: string;
	    anyDeskPath: string;

	    static createFrom(source: any = {}) {
	        return new SettingsInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceLabel = source["deviceLabel"];
	        this.targetAddress = source["targetAddress"];
	        this.tailscalePath = source["tailscalePath"];
	        this.anyDeskPath = source["anyDeskPath"];
	    }
	}
	export class Status {
	    configured: boolean;
	    vpnState: string;
	    targetState: string;
	    anyDeskState: string;
	    message: string;
	    updatedAt: string;
	    settings: SettingsInput;

	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configured = source["configured"];
	        this.vpnState = source["vpnState"];
	        this.targetState = source["targetState"];
	        this.anyDeskState = source["anyDeskState"];
	        this.message = source["message"];
	        this.updatedAt = source["updatedAt"];
	        this.settings = this.convertValues(source["settings"], SettingsInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
