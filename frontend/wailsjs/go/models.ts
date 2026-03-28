export namespace discovery {
	
	export class DiscoveredItem {
	    id: string;
	    label: string;
	    sourcePath: string;
	    found: boolean;
	    partial: string[];
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.sourcePath = source["sourcePath"];
	        this.found = source["found"];
	        this.partial = source["partial"];
	    }
	}
	export class Result {
	    items: DiscoveredItem[];
	    foundCount: number;
	    totalCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], DiscoveredItem);
	        this.foundCount = source["foundCount"];
	        this.totalCount = source["totalCount"];
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

export namespace main {
	
	export class ItemView {
	    id: string;
	    name: string;
	    description: string;
	    notes: string;
	    strategy: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.notes = source["notes"];
	        this.strategy = source["strategy"];
	    }
	}
	export class CategoryView {
	    name: string;
	    items: ItemView[];
	
	    static createFrom(source: any = {}) {
	        return new CategoryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.items = this.convertValues(source["items"], ItemView);
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
	export class ProfileView {
	    name: string;
	    ids: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProfileView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ids = source["ids"];
	    }
	}
	export class ConfigView {
	    backupRoot: string;
	    categories: CategoryView[];
	    profiles: ProfileView[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backupRoot = source["backupRoot"];
	        this.categories = this.convertValues(source["categories"], CategoryView);
	        this.profiles = this.convertValues(source["profiles"], ProfileView);
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
	export class DetectResultView {
	    itemId: string;
	    destPath: string;
	    method: string;
	    confirmed: boolean;
	    candidates: string[];
	
	    static createFrom(source: any = {}) {
	        return new DetectResultView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.destPath = source["destPath"];
	        this.method = source["method"];
	        this.confirmed = source["confirmed"];
	        this.candidates = source["candidates"];
	    }
	}
	export class FolderEntry {
	    name: string;
	    path: string;
	    isDir: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new FolderEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	    }
	}
	
	export class PreflightItem {
	    id: string;
	    label: string;
	    path: string;
	    found: boolean;
	    destPath?: string;
	    destOK: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PreflightItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.path = source["path"];
	        this.found = source["found"];
	        this.destPath = source["destPath"];
	        this.destOK = source["destOK"];
	    }
	}
	export class PreflightResult {
	    sourceRootOK: boolean;
	    destRootOK: boolean;
	    hasItemWarnings: boolean;
	    items: PreflightItem[];
	    readyCount: number;
	    totalCount: number;
	
	    static createFrom(source: any = {}) {
	        return new PreflightResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceRootOK = source["sourceRootOK"];
	        this.destRootOK = source["destRootOK"];
	        this.hasItemWarnings = source["hasItemWarnings"];
	        this.items = this.convertValues(source["items"], PreflightItem);
	        this.readyCount = source["readyCount"];
	        this.totalCount = source["totalCount"];
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

