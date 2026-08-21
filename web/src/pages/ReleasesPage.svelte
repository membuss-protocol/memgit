<script>
  import { Tag, ShieldCheck, Download, Plus, Copy, Check, Calendar, Globe, X } from 'lucide-svelte';
  import { api } from '../api/client.js';

  let { repoName } = $props();

  let releases = $state([]);
  let showDraftModal = $state(false);
  let tagName = $state('');
  let title = $state('');
  let description = $state('');
  let loading = $state(false);
  let copiedMID = $state('');

  $effect(() => {
    loadReleases();
  });

  async function loadReleases() {
    try {
      releases = await api.getReleases(repoName);
    } catch (e) {
      console.error(e);
    }
  }

  function copyMID(mid) {
    navigator.clipboard.writeText(mid);
    copiedMID = mid;
    setTimeout(() => {
      copiedMID = '';
    }, 2000);
  }

  async function handleDraftRelease(e) {
    e.preventDefault();
    if (!tagName.trim()) return;
    loading = true;
    try {
      await api.createRelease(repoName, {
        tag_name: tagName.trim(),
        title: title.trim() || tagName.trim(),
        description: description.trim(),
      });
      tagName = '';
      title = '';
      description = '';
      showDraftModal = false;
      await loadReleases();
    } catch (err) {
      alert(err.message);
    } finally {
      loading = false;
    }
  }
</script>

<div class="releases-page">
  <div class="flex items-center justify-between mb-6">
    <div>
      <h3 class="text-lg font-bold text-main font-mono">Immutable Releases & Merkle DAG Snapshots</h3>
      <p class="text-xs text-muted">Each release captures a permanent, content-addressed MID with 10+4 Reed-Solomon erasure coding.</p>
    </div>

    <button class="btn btn-primary btn-sm" onclick={() => showDraftModal = true}>
      <Plus class="w-3.5 h-3.5" />
      <span>Draft Release</span>
    </button>
  </div>

  <div class="space-y-6">
    {#if !releases || releases.length === 0}
      <div class="card p-12 text-center text-muted text-sm">
        No releases yet. Draft your first release to freeze a permanent Merkle DAG snapshot MID.
      </div>
    {:else}
      {#each releases as release}
        <div class="card card-elevated p-6">
          <div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-4 pb-4 border-b border-slate-800">
            <div class="flex items-center gap-3">
              <span class="badge badge-emerald text-sm font-mono px-3 py-1">
                <Tag class="w-3.5 h-3.5" />
                <span>{release.tag_name}</span>
              </span>
              <h4 class="text-base font-bold text-main">{release.title}</h4>
            </div>

            <span class="text-xs text-dim font-mono">
              Released {new Date(release.created_at).toLocaleDateString()}
            </span>
          </div>

          <div class="text-sm text-muted whitespace-pre-wrap mb-6">
            {release.description || 'No release notes provided.'}
          </div>

          <!-- Root MID Archival Pill -->
          {#if release.mid}
            <div class="p-3.5 bg-slate-900/80 border border-slate-800 rounded-lg flex flex-col md:flex-row items-start md:items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <ShieldCheck class="w-4 h-4 text-emerald-400 shrink-0" />
                <span class="text-xs text-dim font-mono">Merkle DAG Root MID:</span>
                <span class="text-xs font-mono text-emerald-300 select-all">{release.mid}</span>
              </div>

              <div class="flex items-center gap-2">
                <button class="btn btn-secondary btn-sm" onclick={() => copyMID(release.mid)}>
                  {#if copiedMID === release.mid}
                    <Check class="w-3.5 h-3.5 text-emerald-400" />
                    <span class="text-emerald-400 text-xs">Copied</span>
                  {:else}
                    <Copy class="w-3.5 h-3.5 text-dim" />
                    <span class="text-xs">Copy MID</span>
                  {/if}
                </button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  <!-- Draft Release Modal -->
  {#if showDraftModal}
    <div
      class="modal-overlay"
      onclick={() => showDraftModal = false}
      onkeydown={(e) => e.key === 'Escape' && (showDraftModal = false)}
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
          <h3 class="text-base font-bold text-main font-mono">Draft New Release</h3>
          <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={() => showDraftModal = false} aria-label="Close modal">
            <X class="w-4 h-4" />
          </button>
        </div>
        <form onsubmit={handleDraftRelease} class="p-6">
          <div class="mb-4">
            <label for="tag-name" class="block text-xs font-semibold text-main mb-1.5 font-mono">Tag Version *</label>
            <input
              id="tag-name"
              type="text"
              bind:value={tagName}
              placeholder="e.g. v1.0.0"
              class="input-field font-mono"
              required
            />
          </div>
          <div class="mb-4">
            <label for="release-title" class="block text-xs font-semibold text-main mb-1.5">Release Title</label>
            <input
              id="release-title"
              type="text"
              bind:value={title}
              placeholder="e.g. Initial Production Release"
              class="input-field"
            />
          </div>
          <div class="mb-6">
            <label for="release-desc" class="block text-xs font-semibold text-main mb-1.5">Release Notes</label>
            <textarea
              id="release-desc"
              bind:value={description}
              rows="4"
              placeholder="Describe the highlights and changelog..."
              class="input-field"
            ></textarea>
          </div>
          <div class="flex justify-end gap-3 pt-4 border-t border-slate-800">
            <button type="button" class="btn btn-secondary text-xs" onclick={() => showDraftModal = false}>Cancel</button>
            <button type="submit" class="btn btn-primary text-xs" disabled={loading}>
              Publish Release Snapshot
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
