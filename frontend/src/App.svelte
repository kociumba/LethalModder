<script>
    // @ts-ignore
    import { onMount } from "svelte";
    import { writable } from "svelte/store";
    import ProfileSelection from "./ProfileSelection.svelte";
    import ProfileEdit from "./ProfileEdit.svelte";
    import { Return10Simple, Return10WithSearch, GetTotalItems } from "../wailsjs/go/main/App";

    let showProfileSelection = true;
    let currentPage = 1;
    let currentIndex = 0;
    const itemsPerPage = 10;
    let listings = [];
    let totalItems = 0;

    export const searchStore = writable(false);
    export const searchTermStore = writable("");

    const Direction = {
        Next: 0,
        Previous: 1
    };

    onMount(async () => {
        console.log(
            "%cHOW DID YOU GET HERE ?",
            "font-size: 3em; color: crimson;",
        );
        await init();
    });

    async function init() {
        totalItems = await getTotalItemsCount();
        await fetchListings(Direction.Next);
    }

    async function fetchListings(direction) {
        try {
            let isSearching;
            let searchTerm;
            
            // Subscribe to the stores to get their current values
            searchStore.subscribe(value => isSearching = value)();
            searchTermStore.subscribe(value => searchTerm = value)();

            if (isSearching) {
                listings = await Return10WithSearch(currentIndex, direction, searchTerm);
            } else {
                listings = await Return10Simple(currentIndex, direction);
            }
            if (direction === Direction.Next) {
                currentIndex += listings.length;
            }
            // We don't update currentIndex for Previous direction here
        } catch (error) {
            console.error("Error fetching listings:", error);
            listings = [];
        }
    }

    async function getTotalItemsCount() {
        try {
            return await GetTotalItems();
        } catch (error) {
            console.error("Error fetching total count:", error);
            return 0;
        }
    }

    // This is funny 💀
    async function changePage(newPage) {
        if (newPage < 1 || newPage > Math.ceil(totalItems / itemsPerPage))
            return;

        const direction = newPage > currentPage ? Direction.Next : Direction.Previous;
        
        if (direction === Direction.Previous) {
            currentIndex = Math.max(0, currentIndex - (itemsPerPage));
        }

        await fetchListings(direction);
        currentPage = newPage;
    }

    function togglePage() {
        showProfileSelection = !showProfileSelection;
    }
</script>

<main class="container" style="overflow: hidden; height: 100vh">
    {#if showProfileSelection}
        <ProfileSelection on:createProfile={togglePage} />
    {:else}
        <ProfileEdit
            {listings}
            {currentPage}
            {totalItems}
            {itemsPerPage}
            {searchStore} 
            {searchTermStore}
            on:changePage={({ detail }) => changePage(detail)}
            on:backToProfiles={togglePage}
        />
    {/if}
    <button on:click={togglePage}></button>
</main>