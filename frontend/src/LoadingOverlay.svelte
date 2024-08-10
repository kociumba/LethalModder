<script>
    export let isLoading = false;
    export let loadingText = "Loading";
</script>

<div class="overlay" class:hidden={!isLoading}>
    {#if isLoading}
        <span aria-busy="true" class="loading-text"
            >{loadingText}, please wait...</span
        >
    {/if}
</div>

<style>
    .overlay {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.7);
        display: flex;
        justify-content: center;
        align-items: center;
        z-index: 1000;
        transition: opacity 0.3s ease;
        backdrop-filter: blur(10px);
    }

    .overlay.hidden {
        opacity: 0;
        pointer-events: none;
    }

    .loading-text {
        color: white;
        font-size: 1.5rem;
    }
</style>
