<script>
    import { createEventDispatcher } from "svelte";
    import LoadingOverlay from "./LoadingOverlay.svelte";
    import { Events } from "@wailsio/runtime";
    import { CreateProfile } from "../bindings/github.com/kociumba/LethalModder/dataservice";

    let loadingText = "Initializing BepInEx";
    let isLoading = false;

    const dispatch = createEventDispatcher();

    function createProfile() {
        const profileName = prompt("Enter profile name:");
        if (profileName) {
            CreateProfile(profileName);
            isLoading = true;
        }
    }

    Events.On("bepinexInstalled", (data) => {
        console.log("bepinexInstalled", data);
        isLoading = !isLoading;
    });
</script>

<div id="profile-selection-page">
    <h2>Select a Profile</h2>
    <button on:click={createProfile}>Create New Profile</button>
</div>

<LoadingOverlay {loadingText} {isLoading} />
