// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {context} from '../models';
import {rod} from '../models';
import {backend} from '../models';
import {version} from '../models';

export function DeleteMatchLog(arg1:string):Promise<void>;

export function DomReady(arg1:context.Context):Promise<void>;

export function ExportLogToCSV(arg1:string):Promise<void>;

export function FetchData(arg1:string,arg2:rod.Page,arg3:boolean):Promise<void>;

export function GetAppVersion():Promise<string>;

export function GetAvailableLogs():Promise<Array<string>>;

export function GetMatchHistory():Promise<backend.MatchHistory>;

export function GetMatchLog(arg1:string):Promise<Array<backend.MatchHistory>>;

export function Initialize(arg1:string,arg2:string,arg3:version.Version):Promise<number>;

export function IsInitialized():Promise<boolean>;

export function IsTracking():Promise<boolean>;

export function OpenResultsDirectory():Promise<void>;

export function ResultsJSONExist():Promise<boolean>;

export function StartTracking(arg1:string,arg2:boolean):Promise<void>;

export function StopTracking():Promise<void>;
