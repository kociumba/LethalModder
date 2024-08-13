<script>
    import { createEventDispatcher } from "svelte";
    import LoadingOverlay from "./LoadingOverlay.svelte";
    import { Events } from "@wailsio/runtime";
    import { CreateProfile, GetProfiles, OpenProfileDirectory, SelectProfile } from "../bindings/github.com/kociumba/LethalModder/dataservice";
    import { Profile } from "../bindings/github.com/kociumba/LethalModder/profiles";

    /**
     * @type {Profile[]}
     */
    export let profiles = [];

    let loadingText = "Initializing BepInEx";
    let isLoading = false;

    const dispatch = createEventDispatcher();

    async function createProfile() {
        const profileName = prompt("Enter profile name:");
        if (profileName) {
            CreateProfile(profileName);
            isLoading = true;
            profiles = await GetProfiles();
        }
    }

    Events.On("bepinexInstalled", (data) => {
        console.log("bepinexInstalled", data);
        isLoading = !isLoading;
    });
</script>

<div id="profile-selection-page">
    <h2>Select a Profile</h2>

    <div class="section" style="max-width: 90%; margin: 0 auto; height: 70vh; overflow: auto">
        <ul>
            {#each profiles as profile}
                <li>
                    <article role="group" class="grid">
                        <span class="listing-name">{profile.name}</span>
                        <div class="button-group">
                            <button on:click={() => SelectProfile(profile)}>Select Profile</button>
                            <button on:click={() => OpenProfileDirectory(profile)}>Open Profile Directory</button>
                        </div>
                    </article>
                </li>
            {/each}
        </ul>
    </div>

    <button on:click={createProfile}>Create New Profile</button>
</div>

<LoadingOverlay {loadingText} {isLoading} />
