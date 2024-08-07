// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {api} from '../models';

export function Download(arg1:string):Promise<string>;

export function FilterMods(arg1:string):Promise<Array<main.SimplePackageListing>>;

export function GetDownloadURL(arg1:api.PackageListing):Promise<string>;

export function GetTotalItems():Promise<number>;

export function GetTotalItemsFiltered():Promise<number>;

export function Return10Listings(arg1:number,arg2:main.Direction):Promise<Array<api.PackageListing>>;

export function Return10Simple(arg1:number,arg2:main.Direction):Promise<Array<main.SimplePackageListing>>;

export function Return10WithSearch(arg1:number,arg2:main.Direction,arg3:string):Promise<Array<main.SimplePackageListing>>;
