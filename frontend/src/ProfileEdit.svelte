<script>
    // @ts-ignore
    import { createEventDispatcher } from "svelte";
    import { Download, GetTotalItemsFiltered } from "../wailsjs/go/main/App";
    import { BrowserOpenURL } from "../wailsjs/runtime/runtime";

    /**
     * @type {{name: string, description: string, url: string, download_url: string, icon: string}[]}
     */
    export let listings = [];
    export let currentPage = 1;
    export let totalItems = 0;
    export let itemsPerPage = 10;
    let isLoading = true;
    let searchTerm = "";
    export let searchStore;
    export let searchTermStore;
    let totalItemsFiltered = 0;

    const dispatch = createEventDispatcher();

    $: lastPage = Math.ceil(totalItems / itemsPerPage);

    function changePage(newPage) {
        dispatch("changePage", newPage);
    }

    async function downloadMod(url) {
        try {
            await Download(url);
        } catch (error) {
            console.error("Error downloading mod:", error);
        }
    }

    function openWebsite(url) {
        BrowserOpenURL(url);
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
</script>

<div id="profile-edit-page">
    <h2>Edit Profile: <span id="profile-name"></span></h2>

    <div id="mods-section" class="section">
        <div
            style="max-width: 90%; margin: 0 auto; height: 70vh; overflow: auto"
        >
            <ul id="mods-list">
                {#each listings as listing}
                {#if listing.name != '' || listing.description != '' || listing.url != '' || listing.download_url != '' || listing.icon != ''}
                    <li>
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
                                                downloadMod(
                                                    listing.download_url,
                                                )}>Download</button
                                        >
                                        <button
                                            on:click={() =>
                                                openWebsite(listing.url)}
                                            >Open website &rarr;</button
                                        >
                                    </div>
                                </summary>
                                <p>{listing.description}</p>
                            </details>
                        </article>
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
