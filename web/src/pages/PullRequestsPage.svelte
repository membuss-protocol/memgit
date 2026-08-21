<script>
  import { GitPullRequest, GitMerge, Check, Plus, ArrowRight, User, Calendar, Loader2, X } from 'lucide-svelte';
  import { api } from '../api/client.js';

  let { repoName } = $props();

  let prs = $state([]);
  let activeTab = $state('open'); // 'open' | 'merged' | 'closed'
  let showNewPRModal = $state(false);
  let branches = $state([]);

  // New PR Form
  let title = $state('');
  let description = $state('');
  let sourceBranch = $state('');
  let targetBranch = $state('main');
  let author = $state('Developer');

  let loading = $state(false);

  $effect(() => {
    loadPRs();
    loadBranches();
  });

  async function loadPRs() {
    try {
      prs = await api.getPRs(repoName);
    } catch (e) {
      console.error(e);
    }
  }

  async function loadBranches() {
    try {
      branches = await api.getBranches(repoName);
      if (branches.length > 0) {
        targetBranch = branches.find((b) => b.is_default)?.name || branches[0].name;
        sourceBranch = branches[0].name;
      }
    } catch (e) {
      console.error(e);
    }
  }

  let filteredPRs = $derived(prs.filter((p) => p.state === activeTab));

  async function handleCreatePR(e) {
    e.preventDefault();
    if (!title.trim()) return;
    loading = true;
    try {
      await api.createPR(repoName, {
        title: title.trim(),
        description: description.trim(),
        source_branch: sourceBranch,
        target_branch: targetBranch,
        author: author.trim() || 'Developer',
        source_repo: repoName,
      });
      title = '';
      description = '';
      showNewPRModal = false;
      await loadPRs();
    } catch (err) {
      alert(err.message);
    } finally {
      loading = false;
    }
  }

  async function handleMerge(id) {
    if (!confirm(`Are you sure you want to merge Pull Request #${id}?`)) return;
    try {
      await api.mergePR(repoName, id);
      await loadPRs();
    } catch (err) {
      alert(err.message);
    }
  }
</script>

<div class="prs-page card card-elevated">
  <!-- Action Bar -->
  <div class="flex items-center justify-between p-4 border-b border-slate-800 bg-slate-900/60">
    <div class="flex items-center gap-2">
      <button
        class="btn btn-sm {activeTab === 'open' ? 'btn-primary' : 'btn-ghost'}"
        onclick={() => activeTab = 'open'}
      >
        <GitPullRequest class="w-3.5 h-3.5" />
        <span>Open ({prs.filter((p) => p.state === 'open').length})</span>
      </button>
      <button
        class="btn btn-sm {activeTab === 'merged' ? 'btn-primary' : 'btn-ghost'}"
        onclick={() => activeTab = 'merged'}
      >
        <GitMerge class="w-3.5 h-3.5 text-purple-400" />
        <span>Merged ({prs.filter((p) => p.state === 'merged').length})</span>
      </button>
    </div>

    <button class="btn btn-primary btn-sm" onclick={() => showNewPRModal = true}>
      <Plus class="w-3.5 h-3.5" />
      <span>New Pull Request</span>
    </button>
  </div>

  <!-- PR Items -->
  <div class="divide-y divide-slate-800">
    {#if filteredPRs.length === 0}
      <div class="p-12 text-center text-muted text-sm">
        No {activeTab} pull requests in this repository.
      </div>
    {:else}
      {#each filteredPRs as pr}
        <div class="p-4 hover:bg-slate-800/40 transition-colors flex items-start justify-between gap-4">
          <div class="flex items-start gap-3">
            {#if pr.state === 'merged'}
              <GitMerge class="w-4 h-4 text-purple-400 mt-1 shrink-0" />
            {:else}
              <GitPullRequest class="w-4 h-4 text-emerald-400 mt-1 shrink-0" />
            {/if}
            <div>
              <div class="flex items-center gap-2 mb-1">
                <h4 class="text-sm font-semibold text-main">{pr.title}</h4>
                <span class="text-xs text-dim font-mono">#{pr.id}</span>
              </div>

              <!-- Branch Flow -->
              <div class="flex items-center gap-2 mb-2 text-xs font-mono">
                <span class="badge badge-cyan">{pr.source_branch}</span>
                <ArrowRight class="w-3 h-3 text-dim" />
                <span class="badge badge-emerald">{pr.target_branch}</span>
              </div>

              <div class="flex items-center gap-2 text-xs text-dim">
                <span>By {pr.author}</span>
                <span>•</span>
                <span>Created {new Date(pr.created_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>

          <!-- Merge Action -->
          {#if pr.state === 'open'}
            <button class="btn btn-primary btn-sm" onclick={() => handleMerge(pr.id)}>
              <GitMerge class="w-3.5 h-3.5" />
              <span>Merge PR</span>
            </button>
          {:else if pr.state === 'merged'}
            <span class="badge badge-purple font-mono">Merged</span>
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  <!-- New PR Modal -->
  {#if showNewPRModal}
    <div
      class="modal-overlay"
      onclick={() => showNewPRModal = false}
      onkeydown={(e) => e.key === 'Escape' && (showNewPRModal = false)}
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
        <div class="p-5 border-b border-slate-800 flex items-center justify-between">
          <h3 class="text-base font-bold text-main font-mono">Create Pull Request</h3>
          <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={() => showNewPRModal = false} aria-label="Close modal">
            <X class="w-4 h-4" />
          </button>
        </div>
        <form onsubmit={handleCreatePR} class="p-6">
          <div class="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label for="pr-source" class="block text-xs font-semibold text-main mb-1.5 font-mono">Source Branch</label>
              <select id="pr-source" bind:value={sourceBranch} class="input-field font-mono text-xs">
                {#each branches as b}
                  <option value={b.name}>{b.name}</option>
                {/each}
              </select>
            </div>
            <div>
              <label for="pr-target" class="block text-xs font-semibold text-main mb-1.5 font-mono">Target Branch</label>
              <select id="pr-target" bind:value={targetBranch} class="input-field font-mono text-xs">
                {#each branches as b}
                  <option value={b.name}>{b.name}</option>
                {/each}
              </select>
            </div>
          </div>
          <div class="mb-4">
            <label for="pr-title" class="block text-xs font-semibold text-main mb-1.5 font-mono">Title *</label>
            <input
              id="pr-title"
              type="text"
              bind:value={title}
              placeholder="e.g. Implement P2P streaming optimizations"
              class="input-field"
              required
            />
          </div>
          <div class="mb-4">
            <label for="pr-desc" class="block text-xs font-semibold text-main mb-1.5">Description</label>
            <textarea
              id="pr-desc"
              bind:value={description}
              rows="3"
              placeholder="Explain the proposed changes..."
              class="input-field"
            ></textarea>
          </div>
          <div class="mb-6">
            <label for="pr-author" class="block text-xs font-semibold text-main mb-1.5 font-mono">Author</label>
            <input
              id="pr-author"
              type="text"
              bind:value={author}
              placeholder="Your name"
              class="input-field font-mono"
            />
          </div>
          <div class="flex justify-end gap-3 pt-4 border-t border-slate-800">
            <button type="button" class="btn btn-secondary text-xs" onclick={() => showNewPRModal = false}>Cancel</button>
            <button type="submit" class="btn btn-primary text-xs" disabled={loading}>
              {#if loading}
                <Loader2 class="w-4 h-4 animate-spin" />
              {:else}
                <GitPullRequest class="w-4 h-4" />
              {/if}
              <span>Open Pull Request</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
