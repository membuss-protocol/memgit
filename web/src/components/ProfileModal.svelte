<script>
  import { X, User, Key, Check, Shield, Save } from 'lucide-svelte';
  import { api } from '../api/client.js';

  let { currentUser, onClose, onProfileUpdated } = $props();

  let username = $state('');
  let displayName = $state('');
  let bio = $state('');
  let email = $state('');
  let saving = $state(false);
  let error = $state('');
  let success = $state(false);

  $effect(() => {
    if (currentUser) {
      username = currentUser.username || '';
      displayName = currentUser.display_name || '';
      bio = currentUser.bio || '';
      email = currentUser.email || '';
    }
  });

  async function handleSave(e) {
    e.preventDefault();
    saving = true;
    error = '';
    success = false;

    try {
      const updated = await api.updateProfile({
        username: username.trim(),
        display_name: displayName.trim(),
        bio: bio.trim(),
        email: email.trim(),
      });
      success = true;
      if (onProfileUpdated) onProfileUpdated(updated);
      setTimeout(() => {
        onClose();
      }, 1000);
    } catch (err) {
      error = err.message || 'Failed to update profile';
    } finally {
      saving = false;
    }
  }
</script>

<div
  class="modal-overlay"
  onclick={onClose}
  onkeydown={(e) => e.key === 'Escape' && onClose()}
  role="presentation"
  tabindex="-1"
>
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="modal-dialog"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
    role="dialog"
    aria-modal="true"
    tabindex="0"
  >
    <!-- Header -->
    <div class="flex items-center justify-between p-5 border-b border-slate-800">
      <div class="flex items-center gap-2">
        <User class="w-5 h-5 text-emerald-400" />
        <h3 class="text-base font-bold text-main font-mono">Developer Identity Profile</h3>
      </div>
      <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={onClose} aria-label="Close modal">
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Form -->
    <form onsubmit={handleSave}>
      <div class="p-6 space-y-4">
        {#if error}
          <div class="p-3 bg-red-950/40 border border-red-800/50 rounded-lg text-xs text-red-400">
            {error}
          </div>
        {/if}
        {#if success}
          <div class="p-3 bg-emerald-950/40 border border-emerald-800/50 rounded-lg text-xs text-emerald-400 flex items-center gap-2">
            <Check class="w-4 h-4" /> Profile saved successfully!
          </div>
        {/if}

        <!-- Public Key Badge -->
        <div class="p-3 bg-slate-900 border border-slate-800 rounded-lg">
          <div class="flex items-center gap-2 text-xs font-semibold text-cyan-400 mb-1">
            <Shield class="w-3.5 h-3.5" /> Ed25519 Cryptographic Identity
          </div>
          <p class="text-xs text-muted font-mono break-all select-all">
            {currentUser?.public_key || 'Generating...'}
          </p>
        </div>

        <!-- Username -->
        <div>
          <label for="profile-username" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Username handle (@)
          </label>
          <input
            id="profile-username"
            type="text"
            bind:value={username}
            class="input-field w-full font-mono"
            placeholder="e.g. alice"
            required
          />
        </div>

        <!-- Display Name -->
        <div>
          <label for="profile-displayname" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Display Name
          </label>
          <input
            id="profile-displayname"
            type="text"
            bind:value={displayName}
            class="input-field w-full"
            placeholder="e.g. Alice Cooper"
          />
        </div>

        <!-- Bio -->
        <div>
          <label for="profile-bio" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Bio
          </label>
          <textarea
            id="profile-bio"
            bind:value={bio}
            rows="3"
            class="input-field w-full"
            placeholder="Tell the Membuss swarm about your projects and skills..."
          ></textarea>
        </div>

        <!-- Email -->
        <div>
          <label for="profile-email" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Email (for Git signatures)
          </label>
          <input
            id="profile-email"
            type="email"
            bind:value={email}
            class="input-field w-full font-mono"
            placeholder="alice@membuss.network"
          />
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 bg-slate-900/50 border-t border-slate-800 flex justify-end gap-2">
        <button type="button" class="btn btn-secondary text-xs" onclick={onClose}>
          Cancel
        </button>
        <button type="submit" class="btn btn-primary text-xs flex items-center gap-1.5" disabled={saving}>
          <Save class="w-3.5 h-3.5" />
          <span>{saving ? 'Saving...' : 'Save Profile'}</span>
        </button>
      </div>
    </form>
  </div>
</div>
