<script>
  import { ArrowLeft, GitCommit, FileDiff, Plus, Minus, User, Calendar } from 'lucide-svelte';

  let { diff, onBack } = $props();

  function formatPatch(patch) {
    if (!patch) return [];
    return patch.split('\n');
  }
</script>

<div class="diff-viewer card card-elevated">
  <!-- Commit Header -->
  <div class="diff-header">
    <div class="flex items-center justify-between gap-4 mb-4">
      <button class="btn btn-ghost btn-sm flex items-center gap-1.5" onclick={onBack}>
        <ArrowLeft class="w-4 h-4 text-emerald-400" />
        <span>Back to Commits</span>
      </button>

      <div class="flex items-center gap-2">
        <span class="badge badge-emerald font-mono text-xs">
          +{diff?.total_additions || 0}
        </span>
        <span class="badge badge-amber font-mono text-xs">
          -{diff?.total_deletions || 0}
        </span>
      </div>
    </div>

    <h2 class="text-lg font-bold text-main mb-2 font-mono">{diff?.message}</h2>

    <div class="flex flex-wrap items-center gap-4 text-xs text-dim font-mono">
      <span class="flex items-center gap-1 text-muted">
        <User class="w-3.5 h-3.5" />
        <span>{diff?.author?.name} &lt;{diff?.author?.email}&gt;</span>
      </span>
      <span>•</span>
      <span class="flex items-center gap-1">
        <Calendar class="w-3.5 h-3.5" />
        <span>{new Date(diff?.timestamp).toLocaleString()}</span>
      </span>
      <span>•</span>
      <span class="text-cyan-400">commit {diff?.sha}</span>
    </div>
  </div>

  <!-- Changed Files List -->
  <div class="diff-files">
    {#if !diff?.files || diff.files.length === 0}
      <div class="p-8 text-center text-muted text-sm">
        No file changes in this commit (root or empty commit).
      </div>
    {:else}
      {#each diff.files as file}
        <div class="file-diff-box card mb-4">
          <div class="file-diff-header flex items-center justify-between">
            <div class="flex items-center gap-2">
              <FileDiff class="w-4 h-4 text-cyan-400" />
              <span class="font-mono text-sm font-semibold text-main">{file.path}</span>
              <span class="badge badge-slate text-xs">{file.status}</span>
            </div>
            <div class="flex items-center gap-2 font-mono text-xs">
              <span class="text-emerald-400">+{file.additions}</span>
              <span class="text-red-400">-{file.deletions}</span>
            </div>
          </div>

          <!-- Patch Code -->
          {#if file.patch}
            <div class="patch-container font-mono text-xs">
              {#each formatPatch(file.patch) as line}
                {#if line.startsWith('+') && !line.startsWith('+++')}
                  <div class="patch-line patch-add">
                    <span class="line-prefix">+</span>
                    <span class="line-text">{line.slice(1)}</span>
                  </div>
                {:else if line.startsWith('-') && !line.startsWith('---')}
                  <div class="patch-line patch-del">
                    <span class="line-prefix">-</span>
                    <span class="line-text">{line.slice(1)}</span>
                  </div>
                {:else if line.startsWith('@@')}
                  <div class="patch-line patch-hunk">
                    <span class="line-prefix"> </span>
                    <span class="line-text">{line}</span>
                  </div>
                {:else}
                  <div class="patch-line patch-context">
                    <span class="line-prefix"> </span>
                    <span class="line-text">{line.startsWith(' ') ? line.slice(1) : line}</span>
                  </div>
                {/if}
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .diff-viewer {
    padding: 0;
    overflow: hidden;
  }

  .diff-header {
    padding: 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .diff-files {
    padding: 1.25rem;
  }

  .file-diff-box {
    padding: 0;
    overflow: hidden;
    background: #090d16;
    border: 1px solid var(--border-subtle);
  }

  .file-diff-header {
    padding: 0.65rem 1rem;
    background: #0f1523;
    border-bottom: 1px solid var(--border-subtle);
  }

  .patch-container {
    background: #060910;
    overflow-x: auto;
    line-height: 1.5;
  }

  .patch-line {
    display: flex;
    padding: 0.1rem 0.75rem;
    white-space: pre;
  }

  .line-prefix {
    width: 1.5rem;
    user-select: none;
    color: var(--text-dim);
  }

  .patch-add {
    background: rgba(34, 197, 94, 0.12);
    color: #4ade80;
  }

  .patch-del {
    background: rgba(244, 63, 94, 0.12);
    color: #fb7185;
  }

  .patch-hunk {
    background: rgba(56, 189, 248, 0.1);
    color: #38bdf8;
    font-weight: 600;
  }

  .patch-context {
    color: #cbd5e1;
  }
</style>
