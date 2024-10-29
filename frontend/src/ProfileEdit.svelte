<script>
    // @ts-ignore
    import { createEventDispatcher } from "svelte";
    import {
        Download,
        GetTotalItemsFiltered,
    } from "../bindings/github.com/kociumba/LethalModder/dataservice";
    import { OpenURL } from "@wailsio/runtime/src/browser";
    import LoadingOverlay from "./LoadingOverlay.svelte";
    import * as runtime from "@wailsio/runtime";
    import { SimplePackageListing } from "../bindings/github.com/kociumba/LethalModder/models";

    
    let isLoading = $state(false);
    let searchTerm = $state("");
    /** @type {{listings?: SimplePackageListing[], currentPage?: number, totalItems?: number, itemsPerPage?: number, searchStore: any, searchTermStore: any}} */
    let {
        listings = [],
        currentPage = 1,
        totalItems = 0,
        itemsPerPage = 10,
        searchStore,
        searchTermStore
    } = $props();
    let totalItemsFiltered = 0;
    let loadingText = $state("Downloading mod");
    let topElement = $state();

    const dispatch = createEventDispatcher();

    let lastPage = $derived(Math.ceil(totalItems / itemsPerPage));

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

    runtime.Events.On("downloadComplete", () => {
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

    // Debug for showing package info in the frontend
    function handleRightClick(event, listing) {
        if (event.button === 2 && event.shiftKey) {
            event.preventDefault();

            /** @type {runtime.Dialogs.MessageDialogOptions} */
            const options = {
                Title: listing.name,
                Message: JSON.stringify(listing, null, 2),
                Buttons: [{ Label: "OK", IsCancel: false, IsDefault: true }],
                Detached: false,
            };
            runtime.Dialogs.Info(options);
        }
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
                        <li
                            oncontextmenu={(event) =>
                                handleRightClick(event, listing)}
                        >
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
                                                    onload={(event) =>
                                                        handleLoad(event)}
                                                />
                                                {listing.name}
                                            </span>
                                            <div class="button-group">
                                                <button
                                                    onclick={() =>
                                                        downloadMod(listing)}
                                                    >Download</button
                                                >
                                                <button
                                                    onclick={() =>
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
                                                    onload={(event) =>
                                                        handleLoad(event)}
                                                />
                                                {listing.name}
                                            </span>
                                            <div class="button-group">
                                                <button
                                                    onclick={() =>
                                                        downloadMod(listing)}
                                                    >Download</button
                                                >
                                                <button
                                                    onclick={() =>
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
            <button onclick={() => changePage(0)} disabled={currentPage === 1}
                >&laquo;</button
            >
            <button
                onclick={() => changePage(currentPage - 1)}
                disabled={currentPage === 1}>&larr;</button
            >
            <span>{currentPage}</span>
            <button
                onclick={() => changePage(currentPage + 1)}
                disabled={currentPage === lastPage}>&rarr;</button
            >
            <button
                onclick={() => changePage(lastPage)}
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
            oninput={() => {
                if (searchTerm === "") {
                    turnOffSearch();
                } else {
                    turnOnSearch();
                }
            }}
            onchange={() => {
                dispatch("changePage", 0);
            }}
        />
    </div>

    <button onclick={() => dispatch("backToProfiles")}>Back to Profiles</button
    >
</div>

<LoadingOverlay {loadingText} {isLoading} />
