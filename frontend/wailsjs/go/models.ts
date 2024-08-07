export namespace api {
	
	export class Version {
	    date_created: string;
	    dependencies: string[];
	    description: string;
	    download_url: string;
	    downloads: number;
	    file_size: number;
	    full_name: string;
	    icon: string;
	    is_active: boolean;
	    name: string;
	    uuid4: string;
	    version_number: string;
	    website_url: string;
	
	    static createFrom(source: any = {}) {
	        return new Version(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date_created = source["date_created"];
	        this.dependencies = source["dependencies"];
	        this.description = source["description"];
	        this.download_url = source["download_url"];
	        this.downloads = source["downloads"];
	        this.file_size = source["file_size"];
	        this.full_name = source["full_name"];
	        this.icon = source["icon"];
	        this.is_active = source["is_active"];
	        this.name = source["name"];
	        this.uuid4 = source["uuid4"];
	        this.version_number = source["version_number"];
	        this.website_url = source["website_url"];
	    }
	}
	export class PackageListing {
	    name: string;
	    full_name: string;
	    owner: string;
	    package_url: string;
	    donation_link: string;
	    date_created: string;
	    date_updated: string;
	    uuid4: string;
	    rating_score: any;
	    is_pinned: any;
	    is_deprecated: any;
	    has_nsfw_content: boolean;
	    categories: any;
	    versions: Version[];
	
	    static createFrom(source: any = {}) {
	        return new PackageListing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.full_name = source["full_name"];
	        this.owner = source["owner"];
	        this.package_url = source["package_url"];
	        this.donation_link = source["donation_link"];
	        this.date_created = source["date_created"];
	        this.date_updated = source["date_updated"];
	        this.uuid4 = source["uuid4"];
	        this.rating_score = source["rating_score"];
	        this.is_pinned = source["is_pinned"];
	        this.is_deprecated = source["is_deprecated"];
	        this.has_nsfw_content = source["has_nsfw_content"];
	        this.categories = source["categories"];
	        this.versions = this.convertValues(source["versions"], Version);
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
	
	export class SimplePackageListing {
	    name: string;
	    description: string;
	    url: string;
	    download_url: string;
	
	    static createFrom(source: any = {}) {
	        return new SimplePackageListing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.url = source["url"];
	        this.download_url = source["download_url"];
	    }
	}

}

