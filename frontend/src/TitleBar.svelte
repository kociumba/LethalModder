<script>
    import { Application, Window } from "@wailsio/runtime";

    let { title = "Lethal Modder" } = $props();
    let isMaximized = $state(false);

    const minimize = async () => {
        Window.Minimise();
    };

    const maximize = async () => {
        isMaximized = await Window.IsMaximised();
        if (isMaximized) {
            Window.UnMaximise();
        } else {
            Window.Maximise();
        }
    };

    const quit = () => {
        Application.Quit();
    };
</script>

<div class="title-bar" style="--wails-draggable: drag">
    <div class="drag-region">
        <div class="app-title">
            <img src="/appicon.png" alt="App Icon" class="app-icon" />
            <span>{title}</span>
        </div>
    </div>
    <div class="window-controls">
        <button class="control-button minimize" onclick={minimize} aria-label="Minimize">
            <svg width="12" height="2" viewBox="0 0 12 2">
                <rect width="12" height="1" fill="currentColor" />
            </svg>
        </button>
        <button class="control-button maximize" onclick={maximize} aria-label="Maximize">
            {#if !isMaximized}
                <svg width="12" height="12" viewBox="0 0 12 12">
                    <path fill="currentColor" d="M3.5,3.5v5h5v-5h-5z M2,2h8v8H2V2z"/>
                </svg>
            {:else}
                <svg width="12" height="12" viewBox="0 0 12 12">
                    <rect width="10" height="10" x="1" y="1" fill="none" stroke="currentColor" />
                </svg>
            {/if}
        </button>
        <button class="control-button close" onclick={quit} aria-label="Close">
            <svg width="12" height="12" viewBox="0 0 12 12">
                <path fill="currentColor" d="M1,1 L11,11 M1,11 L11,1" stroke="currentColor" stroke-width="1.5"/>
            </svg>
        </button>
    </div>
</div>

<style>
    .title-bar {
        height: 32px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: rgba(255, 255, 255, 0.1);
        backdrop-filter: blur(10px);
        -webkit-backdrop-filter: blur(10px);
        box-shadow: #0000003A 0px 0px 10px 0px;
        user-select: none;
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        z-index: 1000;
    }

    .drag-region {
        flex: 1;
        -webkit-app-region: drag;
        display: flex;
        align-items: center;
        height: 100%;
    }

    .app-title {
        display: flex;
        align-items: center;
        gap: 8px;
        padding-left: 12px;
        color: #ffffff;
        font-size: 14px;
    }

    .app-icon {
        width: 16px;
        height: 16px;
    }

    .window-controls {
        display: flex;
        height: 100%;
        -webkit-app-region: no-drag;
    }

    /* Reset all Pico CSS button styles */
    .control-button {
        all: unset;
        width: 46px;
        height: 100%;
        padding: 0;
        margin: 0;
        background: transparent;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #ffffff;
        transition: background-color 0.1s;
        border: none;
        border-radius: 0;
        box-shadow: none;
        font-size: inherit;
        line-height: 1;
        text-transform: none;
        font-weight: normal;
        letter-spacing: normal;
    }

    /* Override Pico's focus styles */
    .control-button:focus {
        outline: none;
        box-shadow: none;
        background: transparent;
        border: none;
    }

    /* Override Pico's hover styles */
    .control-button:hover {
        background: rgba(255, 255, 255, 0.1);
        transform: none;
        box-shadow: none;
        border: none;
    }

    .close:hover {
        background: #e81123;
    }

    /* Override Pico's active styles */
    .control-button:active {
        background: rgba(255, 255, 255, 0.2);
        transform: none;
        box-shadow: none;
        border: none;
    }

    .close:active {
        background: #f1707a;
    }

    /* Additional overrides for Pico's button states */
    /* .control-button:is([aria-invalid], :invalid) {
        box-shadow: none;
        border: none;
    }

    .control-button[aria-busy="true"] {
        background: transparent;
    }

    .control-button:is([aria-invalid], :invalid):is(:hover, :focus) {
        background: rgba(255, 255, 255, 0.1);
    } */
</style>