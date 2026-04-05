export namespace main {
	
	export class AppSettings {
	    unblockWaiting: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.unblockWaiting = source["unblockWaiting"];
	    }
	}

}

