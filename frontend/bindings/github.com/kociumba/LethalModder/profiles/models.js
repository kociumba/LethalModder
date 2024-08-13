// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import {Create as $Create} from "@wailsio/runtime";

export class Profile {
    /**
     * Creates a new Profile instance.
     * @param {Partial<Profile>} [$$source = {}] - The source object to create the Profile.
     */
    constructor($$source = {}) {
        if (!("name" in $$source)) {
            /**
             * @member
             * @type {string}
             */
            this["name"] = "";
        }
        if (!("path" in $$source)) {
            /**
             * @member
             * @type {string}
             */
            this["path"] = "";
        }
        if (!("mods" in $$source)) {
            /**
             * @member
             * @type {string[]}
             */
            this["mods"] = [];
        }

        Object.assign(this, $$source);
    }

    /**
     * Creates a new Profile instance from a string or object.
     * @param {any} [$$source = {}]
     * @returns {Profile}
     */
    static createFrom($$source = {}) {
        const $$createField2_0 = $$createType0;
        let $$parsedSource = typeof $$source === 'string' ? JSON.parse($$source) : $$source;
        if ("mods" in $$parsedSource) {
            $$parsedSource["mods"] = $$createField2_0($$parsedSource["mods"]);
        }
        return new Profile(/** @type {Partial<Profile>} */($$parsedSource));
    }
}

// Private type creation functions
const $$createType0 = $Create.Array($Create.Any);
