export namespace models {
	
	export class ClipboardItem {
	    id: string;
	    contentType: string;
	    content: string;
	    preview: string;
	    isPinned: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    lastAccessed: any;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.contentType = source["contentType"];
	        this.content = source["content"];
	        this.preview = source["preview"];
	        this.isPinned = source["isPinned"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.lastAccessed = this.convertValues(source["lastAccessed"], null);
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
	export class Settings {
	    id: number;
	    globalHotkey: string;
	    previousItemHotkey: string;
	    pollingInterval: number;
	    maxItems: number;
	    maxDays: number;
	    autoLaunch: boolean;
	    enableSounds: boolean;
	    monitoringEnabled: boolean;
	    allowPasswords: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.globalHotkey = source["globalHotkey"];
	        this.previousItemHotkey = source["previousItemHotkey"];
	        this.pollingInterval = source["pollingInterval"];
	        this.maxItems = source["maxItems"];
	        this.maxDays = source["maxDays"];
	        this.autoLaunch = source["autoLaunch"];
	        this.enableSounds = source["enableSounds"];
	        this.monitoringEnabled = source["monitoringEnabled"];
	        this.allowPasswords = source["allowPasswords"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

