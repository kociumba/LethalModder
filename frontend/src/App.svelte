<script>
    // @ts-ignore
    import { onMount } from "svelte";
    import ProfileSelection from "./ProfileSelection.svelte";
    import ProfileEdit from "./ProfileEdit.svelte";
    import { Return10Simple, GetTotalItems } from "../wailsjs/go/main/App";

    let showProfileSelection = true;
    let currentPage = 1;
    let currentIndex = 0;
    const itemsPerPage = 10;
    let listings = [];
    let totalItems = 0;

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
            listings = await Return10Simple(currentIndex, direction);
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
            on:changePage={({ detail }) => changePage(detail)}
            on:backToProfiles={togglePage}
        />
    {/if}
    <button on:click={togglePage}></button>
</main>