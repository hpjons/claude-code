export namespace main {
	
	export class EvaluationResult {
	    input: string;
	    result: string;
	    isError: boolean;
	    errorMsg?: string;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = source["input"];
	        this.result = source["result"];
	        this.isError = source["isError"];
	        this.errorMsg = source["errorMsg"];
	    }
	}

}

