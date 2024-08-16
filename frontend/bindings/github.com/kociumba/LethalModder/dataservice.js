// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import {Call as $Call, Create as $Create} from "@wailsio/runtime";

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import * as api$0 from "./api/models.js";
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import * as profiles$0 from "./profiles/models.js";

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import * as $models from "./models.js";

/**
 * @param {string} name
 * @returns {Promise<void> & { cancel(): void }}
 */
export function CreateProfile(name) {
    let $resultPromise = /** @type {any} */($Call.ByID(4208716454, name));
    return $resultPromise;
}

/**
 * overcomplicated, r2modman just extracts everything from the bepinex
 * @param {$models.SimplePackageListing} listing
 * @returns {Promise<string> & { cancel(): void }}
 */
export function Download(listing) {
    let $resultPromise = /** @type {any} */($Call.ByID(3764793721, listing));
    return $resultPromise;
}

/**
 * Unsung hero of the search function xd
 * @param {string} search
 * @returns {Promise<$models.SimplePackageListing[]> & { cancel(): void }}
 */
export function FilterMods(search) {
    let $resultPromise = /** @type {any} */($Call.ByID(2555135026, search));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType1($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * [0] is the newest
 * Deprecated as I already return this in the simplified package listing
 * 
 * Deprecated: Use SimplePackageListing.DownloadURL
 * @param {api$0.PackageListing} listing
 * @returns {Promise<string> & { cancel(): void }}
 */
export function GetDownloadURL(listing) {
    let $resultPromise = /** @type {any} */($Call.ByID(4145055636, listing));
    return $resultPromise;
}

/**
 * shitass function, still don't know why the event doesn't get picked up
 * maby multi window stuff
 * @returns {Promise<boolean> & { cancel(): void }}
 */
export function GetIsLethalCompanyInstalled() {
    let $resultPromise = /** @type {any} */($Call.ByID(825651866));
    return $resultPromise;
}

/**
 * @returns {Promise<profiles$0.Profile[]> & { cancel(): void }}
 */
export function GetProfiles() {
    let $resultPromise = /** @type {any} */($Call.ByID(3635952993));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType3($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * @returns {Promise<number> & { cancel(): void }}
 */
export function GetTotalItems() {
    let $resultPromise = /** @type {any} */($Call.ByID(3228312929));
    return $resultPromise;
}

/**
 * @returns {Promise<number> & { cancel(): void }}
 */
export function GetTotalItemsFiltered() {
    let $resultPromise = /** @type {any} */($Call.ByID(591123730));
    return $resultPromise;
}

/**
 * InitializePackageMap should be called once when initializing the DataService
 * @returns {Promise<void> & { cancel(): void }}
 */
export function InitializePackageMap() {
    let $resultPromise = /** @type {any} */($Call.ByID(2403257157));
    return $resultPromise;
}

/**
 * @param {profiles$0.Profile} profile
 * @returns {Promise<boolean> & { cancel(): void }}
 */
export function IsBepInExInstalled(profile) {
    let $resultPromise = /** @type {any} */($Call.ByID(1631519816, profile));
    return $resultPromise;
}

/**
 * Windows only
 * gonna have to make a system check for this, when linux support is going to come
 * @param {profiles$0.Profile} profile
 * @returns {Promise<void> & { cancel(): void }}
 */
export function OpenProfileDirectory(profile) {
    let $resultPromise = /** @type {any} */($Call.ByID(3211936441, profile));
    return $resultPromise;
}

/**
 * Turns out the data is so big even on 10 entries that it crashed webview2 bridge
 * 
 * # Do not use from frontend, results in a stack overflow
 * @param {number} currentIndex
 * @param {$models.Direction} direction
 * @returns {Promise<api$0.PackageListing[]> & { cancel(): void }}
 */
export function Return10Listings(currentIndex, direction) {
    let $resultPromise = /** @type {any} */($Call.ByID(4172540335, currentIndex, direction));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType5($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * @param {number} currentIndex
 * @param {$models.Direction} direction
 * @returns {Promise<$models.SimplePackageListing[]> & { cancel(): void }}
 */
export function Return10Simple(currentIndex, direction) {
    let $resultPromise = /** @type {any} */($Call.ByID(3277498868, currentIndex, direction));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType1($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * @param {number} currentIndex
 * @param {$models.Direction} direction
 * @param {string} search
 * @returns {Promise<$models.SimplePackageListing[]> & { cancel(): void }}
 */
export function Return10WithSearch(currentIndex, direction, search) {
    let $resultPromise = /** @type {any} */($Call.ByID(4219935970, currentIndex, direction, search));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType1($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * @param {profiles$0.Profile} profile
 * @returns {Promise<void> & { cancel(): void }}
 */
export function SelectProfile(profile) {
    let $resultPromise = /** @type {any} */($Call.ByID(328192790, profile));
    return $resultPromise;
}

// Private type creation functions
const $$createType0 = $models.SimplePackageListing.createFrom;
const $$createType1 = $Create.Array($$createType0);
const $$createType2 = profiles$0.Profile.createFrom;
const $$createType3 = $Create.Array($$createType2);
const $$createType4 = api$0.PackageListing.createFrom;
const $$createType5 = $Create.Array($$createType4);
