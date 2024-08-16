<script>
    // @ts-ignore
    import { createEventDispatcher } from "svelte";
    import {
        Download,
        GetTotalItemsFiltered,
    } from "../bindings/github.com/kociumba/LethalModder/dataservice";
    import { OpenURL } from "@wailsio/runtime/src/browser";
    import LoadingOverlay from "./LoadingOverlay.svelte";
    import { Events } from "@wailsio/runtime";
    import { SimplePackageListing } from "../bindings/github.com/kociumba/LethalModder/models";

    /**
     * @type {SimplePackageListing[]}
     *
     * Simplified version of PackageListing, required to avoid a stack overflow in the webview2 bridge
     *
     * This holds 10 currently displayed simple listings
     */
    export let listings = [];
    export let currentPage = 1;
    export let totalItems = 0;
    export let itemsPerPage = 10;
    let isLoading = false;
    let searchTerm = "";
    export let searchStore;
    export let searchTermStore;
    let totalItemsFiltered = 0;
    let loadingText = "Downloading mod";
    let topElement;

    const dispatch = createEventDispatcher();

    $: lastPage = Math.ceil(totalItems / itemsPerPage);

    function changePage(newPage) {
        scrollToTop();
        dispatch("changePage", newPage);
    }

    // damn the wails v3 event are good
    /**
     * @param {SimplePackageListing} listing
     */
    async function downloadMod(listing) {
        loadingText = "Downloading " + listing.name + "_" + listing.version;
        isLoading = true;
        try {
            await Download(listing);
        } catch (error) {
            console.error("Error downloading mod:", error);
            isLoading = false;
        }
    }

    Events.On("downloadComplete", () => {
        isLoading = false;
    });

    function openWebsite(url) {
        OpenURL(url);
    }

    function handleLoad(event) {
        event.target.setAttribute("aria-busy", "false");
    }

    function turnOffSearch() {
        searchStore.set(false);
        searchTermStore.set("");
        searchTerm = "";
    }

    async function turnOnSearch() {
        searchStore.set(true);
        searchTermStore.set(searchTerm);
        totalItemsFiltered = await GetTotalItemsFiltered();
    }

    function canGoForward() {
        if (currentPage === lastPage) {
            return true;
        } else if (searchTerm !== "") {
            return !(currentPage * itemsPerPage < totalItemsFiltered);
        } else {
            return false;
        }
    }

    function scrollToTop() {
        console.log("scrollToTop", topElement);
        topElement.scrollIntoView({ behavior: "smooth" });
    }
</script>

<div id="profile-edit-page">
    <h2>Edit Profile: <span id="profile-name"></span></h2>

    <div id="mods-section" class="section">
        <div
            style="max-width: 90%; margin: 0 auto; height: 70vh; overflow: auto"
        >
            <ul id="mods-list">
                {#each listings as listing, i}
                    {#if listing.name != "" || listing.description != "" || listing.url != "" || listing.download_url != "" || listing.icon != ""}
                        <li>
                            {#if i === 0}
                                <article bind:this={topElement}>
                                    <details>
                                        <summary class="outline contrast">
                                            <span class="listing-name">
                                                <img
                                                    src={listing.icon}
                                                    alt="mod icon"
                                                    height="64"
                                                    width="64"
                                                    aria-busy="true"
                                                    on:load={(event) =>
                                                        handleLoad(event)}
                                                />
                                                {listing.name}
                                            </span>
                                            <div class="button-group">
                                                <button
                                                    on:click={() =>
                                                        downloadMod(listing)}
                                                    >Download</button
                                                >
                                                <button
                                                    on:click={() =>
                                                        openWebsite(
                                                            listing.url,
                                                        )}
                                                    >Open website &rarr;</button
                                                >
                                            </div>
                                        </summary>
                                        <p>{listing.description}</p>
                                    </details>
                                </article>
                            {:else}
                                <article>
                                    <details>
                                        <summary class="outline contrast">
                                            <span class="listing-name">
                                                <img
                                                    src={listing.icon}
                                                    alt="mod icon"
                                                    height="64"
                                                    width="64"
                                                    aria-busy="true"
                                                    on:load={(event) =>
                                                        handleLoad(event)}
                                                />
                                                {listing.name}
                                            </span>
                                            <div class="button-group">
                                                <button
                                                    on:click={() =>
                                                        downloadMod(listing)}
                                                    >Download</button
                                                >
                                                <button
                                                    on:click={() =>
                                                        openWebsite(
                                                            listing.url,
                                                        )}
                                                    >Open website &rarr;</button
                                                >
                                            </div>
                                        </summary>
                                        <p>{listing.description}</p>
                                    </details>
                                </article>
                            {/if}
                        </li>
                    {/if}
                {/each}
            </ul>
        </div>

        <div class="pagination">
            <button on:click={() => changePage(0)} disabled={currentPage === 1}
                >&laquo;</button
            >
            <button
                on:click={() => changePage(currentPage - 1)}
                disabled={currentPage === 1}>&larr;</button
            >
            <span>{currentPage}</span>
            <button
                on:click={() => changePage(currentPage + 1)}
                disabled={currentPage === lastPage}>&rarr;</button
            >
            <button
                on:click={() => changePage(lastPage)}
                disabled={currentPage === lastPage}>&raquo;</button
            >
        </div>
        <input
            type="text"
            name="text"
            placeholder="Text"
            aria-label="Text"
            id="search"
            bind:value={searchTerm}
            on:input={() => {
                if (searchTerm === "") {
                    turnOffSearch();
                } else {
                    turnOnSearch();
                }
            }}
            on:change={() => {
                dispatch("changePage", 0);
            }}
        />
    </div>

    <button on:click={() => dispatch("backToProfiles")}>Back to Profiles</button
    >
</div>

<LoadingOverlay {loadingText} {isLoading} />
