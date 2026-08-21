<script>
  import { X, GitBranch, Plus, Shield, Terminal } from 'lucide-svelte';

  let { onClose, onCreate } = $props();

  let name = $state('');
  let description = $state('');
  let defaultBranch = $state('main');
  let isPrivate = $state(false);
  let initReadme = $state(true);
  let loading = $state(false);
  let error = $state('');

  function handleSubmit(e) {
    e.preventDefault();
    if (!name.trim()) {
      error = 'Repository name is required';
      return;
    }
    // Clean name
    const cleaned = name.trim().toLowerCase().replace(/[^a-z0-9-_]/g, '-');
    loading = true;
    error = '';

    onCreate({
      name: cleaned,
      description: description.trim(),
      default_branch: defaultBranch.trim() || 'main',
      is_private: isPrivate,
      init_readme: initReadme,
    });
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
        <GitBranch class="w-5 h-5 text-emerald-400" />
        <h3 class="text-base font-bold text-main font-mono">Create New Repository</h3>
      </div>
      <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={onClose} aria-label="Close modal">
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Form -->
    <form onsubmit={handleSubmit}>
      <div class="p-6 space-y-4">
        {#if error}
          <div class="p-3 bg-red-950/40 border border-red-800/50 rounded-lg text-xs text-red-400">
            {error}
          </div>
        {/if}

        <!-- Repository Name -->
        <div>
          <label for="new-repo-name" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Repository Name <span class="text-emerald-400">*</span>
          </label>
          <input
            id="new-repo-name"
            type="text"
            bind:value={name}
            class="input-field w-full font-mono text-sm"
            placeholder="e.g. quantum-protocol"
            required
            autocomplete="off"
          />
          <span class="text-[11px] text-muted mt-1 block">
            Will be addressed via MemNS identifier: <span class="font-mono text-cyan-400">/memns/memns1z{name ? name.toLowerCase().replace(/[^a-z0-9-_]/g, '-') : '...'}</span>
          </span>
        </div>

        <!-- Description -->
        <div>
          <label for="new-repo-desc" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Description
          </label>
          <textarea
            id="new-repo-desc"
            bind:value={description}
            rows="3"
            class="input-field w-full"
            placeholder="Describe your decentralized project..."
          ></textarea>
        </div>

        <!-- Default Branch -->
        <div>
          <label for="new-repo-branch" class="block text-xs font-semibold text-muted uppercase tracking-wider mb-1.5">
            Default Branch
          </label>
          <input
            id="new-repo-branch"
            type="text"
            bind:value={defaultBranch}
            class="input-field w-full font-mono"
            placeholder="main"
          />
        </div>

        <!-- Checkbox Options -->
        <div class="pt-2 space-y-3">
          <label class="flex items-center gap-2.5 cursor-pointer text-xs select-none">
            <input
              type="checkbox"
              bind:checked={initReadme}
              class="w-4 h-4 rounded border-slate-700 bg-slate-900 text-emerald-500 focus:ring-emerald-500 focus:ring-offset-slate-900"
            />
            <span class="text-main font-medium">Initialize with README.md</span>
          </label>

          <label class="flex items-center gap-2.5 cursor-pointer text-xs select-none">
            <input
              type="checkbox"
              bind:checked={isPrivate}
              class="w-4 h-4 rounded border-slate-700 bg-slate-900 text-emerald-500 focus:ring-emerald-500 focus:ring-offset-slate-900"
            />
            <span class="text-main font-medium">Private Repository (Encrypted DAG)</span>
          </label>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 bg-slate-900/50 border-t border-slate-800 flex justify-end gap-2">
        <button type="button" class="btn btn-secondary text-xs" onclick={onClose}>
          Cancel
        </button>
        <button type="submit" class="btn btn-primary text-xs flex items-center gap-1.5" disabled={loading}>
          <Plus class="w-3.5 h-3.5" />
          <span>{loading ? 'Creating...' : 'Create Repository'}</span>
        </button>
      </div>
    </form>
  </div>
</div>
