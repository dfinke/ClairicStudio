export namespace clairic {
	
	export class CodeExport {
	    language: string;
	    filename: string;
	    code: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new CodeExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.filename = source["filename"];
	        this.code = source["code"];
	        this.note = source["note"];
	    }
	}
	export class Field {
	    name: string;
	    type: string;
	    instructions: string;
	    settings: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Field(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.instructions = source["instructions"];
	        this.settings = source["settings"];
	    }
	}
	export class Result {
	    answers: Record<string, Array<number>>;
	    truncated: boolean;
	    state_text?: string;
	    usage?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.answers = source["answers"];
	        this.truncated = source["truncated"];
	        this.state_text = source["state_text"];
	        this.usage = source["usage"];
	    }
	}
	export class Schema {
	    version: number;
	    fields: Field[];
	    reference_date?: string;
	    include_quantiles?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Schema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.fields = this.convertValues(source["fields"], Field);
	        this.reference_date = source["reference_date"];
	        this.include_quantiles = source["include_quantiles"];
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
	
	export class Connection {
	    account: string;
	    configured: boolean;
	    source: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.account = source["account"];
	        this.configured = source["configured"];
	        this.source = source["source"];
	        this.message = source["message"];
	    }
	}
	export class Workspace {
	    schema: clairic.Schema;
	    theme: string;
	    document_path: string;
	    schema_path: string;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema = this.convertValues(source["schema"], clairic.Schema);
	        this.theme = source["theme"];
	        this.document_path = source["document_path"];
	        this.schema_path = source["schema_path"];
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
	export class Bootstrap {
	    workspace: Workspace;
	    document: string;
	    connection: Connection;
	    notice: string;
	
	    static createFrom(source: any = {}) {
	        return new Bootstrap(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspace = this.convertValues(source["workspace"], Workspace);
	        this.document = source["document"];
	        this.connection = this.convertValues(source["connection"], Connection);
	        this.notice = source["notice"];
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
	
	export class Document {
	    path: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Document(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.text = source["text"];
	    }
	}
	export class SchemaFile {
	    path: string;
	    schema: clairic.Schema;
	
	    static createFrom(source: any = {}) {
	        return new SchemaFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.schema = this.convertValues(source["schema"], clairic.Schema);
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

