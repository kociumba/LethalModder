<script>
    export let isWarningVisible = false;
    export let warningText = `
    Lethal Company is not installed.
    Please install it through Steam.
    `;
</script>

<div class="overlay" class:hidden={!isWarningVisible}>
    {#if isWarningVisible}
        <div class="warning-message">
            <span class="warning-icon">⚠️</span>
            <span class="warning-text">{warningText}</span>
        </div>
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

    .warning-message {
        display: flex;
        align-items: center;
        color: white;
        font-size: 1.5rem;
    }

    .warning-icon {
        margin-right: 10px;
    }

    .warning-text {
        font-size: 1.5rem;
    }
</style>
