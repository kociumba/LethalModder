<script>
    import { createEventDispatcher } from 'svelte';
  
    const dispatch = createEventDispatcher();
  
    function createProfile() {
      const profileName = prompt('Enter profile name:');
      if (profileName) {
        // Here you would typically handle profile creation logic
        dispatch('createProfile', { name: profileName });
      }
    }
  </script>
  
  <div id="profile-selection-page">
    <h2>Select a Profile</h2>
    <button on:click={createProfile}>Create New Profile</button>
  </div>