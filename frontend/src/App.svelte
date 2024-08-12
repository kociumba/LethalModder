<script>
    import { onMount } from "svelte";
    import { writable } from "svelte/store";
    import ProfileSelection from "./ProfileSelection.svelte";
    import ProfileEdit from "./ProfileEdit.svelte";
    import LethalCompanyWarningOverlay from "./LethalCompanyWarningOverlay.svelte";
    import { 
        Return10Simple, 
        Return10WithSearch, 
        GetTotalItems, 
        GetIsLethalCompanyInstalled,
    } from "../bindings/github.com/kociumba/LethalModder/dataservice";
    import { Events } from "@wailsio/runtime"

    let showProfileSelection = true;
    let currentPage = 1;
    let currentIndex = 0;
    const itemsPerPage = 10;
    let listings = [];
    let totalItems = 0;
    let isWarningVisible = false;

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

    // actually breaks without this 🤷
    async function init() {
        totalItems = await GetTotalItems();
        await fetchListings(Direction.Next);
        isWarningVisible = !(await GetIsLethalCompanyInstalled());
    }

    // Needed to get totalItems after they loaded
    Events.On("totalItems", async function(data) {
        totalItems = data
        await fetchListings(Direction.Next);
    })

    Events.On("lethalCompanyWarning", async function(data) {
        isWarningVisible = !data
        console.log("isWarningVisible", isWarningVisible)
    })

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

    async function fetchListingsManual(index, direction) {
        try {
            let isSearching;
            let searchTerm;
            
            // Subscribe to the stores to get their current values
            searchStore.subscribe(value => isSearching = value)();
            searchTermStore.subscribe(value => searchTerm = value)();

            if (isSearching) {
                listings = await Return10WithSearch(index, direction, searchTerm);
            } else {
                listings = await Return10Simple(index, direction);
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
        if (newPage === 0) {
            fetchListingsManual(0, Direction.Next);
            currentIndex = 0;
            currentPage = 1;
            return;
        }

        if (newPage < 1 || newPage > Math.ceil(totalItems / itemsPerPage))
            return;

        const direction = newPage > currentPage ? Direction.Next : Direction.Previous;
        
        if (direction === Direction.Previous) {
            currentIndex = Math.max(0, currentIndex - (itemsPerPage));
        }

        await fetchListings(direction);
        currentPage = newPage;

        console.log("isWarningVisible", isWarningVisible)
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

<LethalCompanyWarningOverlay {isWarningVisible}/>