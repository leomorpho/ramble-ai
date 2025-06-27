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
	    }
	}

}

