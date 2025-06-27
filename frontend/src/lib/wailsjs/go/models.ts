export namespace main {
	
	export class LocalVideoFile {
	    name: string;
	    filePath: string;
	    fileName: string;
	    fileSize: number;
	    format: string;
	    exists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LocalVideoFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.filePath = source["filePath"];
	        this.fileName = source["fileName"];
	        this.fileSize = source["fileSize"];
	        this.format = source["format"];
	        this.exists = source["exists"];
	    }
	}
	export class ProjectResponse {
	    id: number;
	    name: string;
	    description: string;
	    path: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.path = source["path"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TestOpenAIApiKeyResponse {
	    valid: boolean;
	    message: string;
	    model?: string;
	
	    static createFrom(source: any = {}) {
	        return new TestOpenAIApiKeyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	        this.model = source["model"];
	    }
	}
	export class Word {
	    word: string;
	    start: number;
	    end: number;
	
	    static createFrom(source: any = {}) {
	        return new Word(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}
	export class TranscriptionResponse {
	    success: boolean;
	    message: string;
	    transcription?: string;
	    words?: Word[];
	    language?: string;
	    duration?: number;
	
	    static createFrom(source: any = {}) {
	        return new TranscriptionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.transcription = source["transcription"];
	        this.words = this.convertValues(source["words"], Word);
	        this.language = source["language"];
	        this.duration = source["duration"];
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
	export class VideoClipResponse {
	    id: number;
	    name: string;
	    description: string;
	    filePath: string;
	    fileName: string;
	    fileSize: number;
	    duration: number;
	    format: string;
	    width: number;
	    height: number;
	    projectId: number;
	    createdAt: string;
	    updatedAt: string;
	    exists: boolean;
	    thumbnailUrl: string;
	    transcription: string;
	    transcriptionWords: Word[];
	    transcriptionLanguage: string;
	    transcriptionDuration: number;
	
	    static createFrom(source: any = {}) {
	        return new VideoClipResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.filePath = source["filePath"];
	        this.fileName = source["fileName"];
	        this.fileSize = source["fileSize"];
	        this.duration = source["duration"];
	        this.format = source["format"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.projectId = source["projectId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.exists = source["exists"];
	        this.thumbnailUrl = source["thumbnailUrl"];
	        this.transcription = source["transcription"];
	        this.transcriptionWords = this.convertValues(source["transcriptionWords"], Word);
	        this.transcriptionLanguage = source["transcriptionLanguage"];
	        this.transcriptionDuration = source["transcriptionDuration"];
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

