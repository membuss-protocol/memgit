<script>
  import { GitCommit, Copy, Check, Calendar, User, ArrowRight } from 'lucide-svelte';

  let { commits, onSelectCommit } = $props();
  let copiedSHA = $state('');

  function copyHash(e, sha) {
    e.stopPropagation();
    navigator.clipboard.writeText(sha);
    copiedSHA = sha;
    setTimeout(() => {
      copiedSHA = '';
    }, 2000);
  }
</script>

<div class="commit-list card card-elevated">
  <div class="commit-header flex items-center justify-between">
    <div class="flex items-center gap-2">
      <GitCommit class="w-4 h-4 text-emerald-400" />
      <span class="font-mono text-sm font-semibold text-main">Commit History</span>
      <span class="badge badge-slate text-xs font-mono">{commits?.length || 0} commits</span>
    </div>
  </div>

  <div class="commit-timeline">
    {#if !commits || commits.length === 0}
      <div class="p-8 text-center text-muted text-sm">
        No commits found on this branch.
      </div>
    {:else}
      {#each commits as commit}
        <div
          class="commit-row cursor-pointer"
          onclick={() => onSelectCommit(commit.sha)}
          role="button"
          tabindex="0"
          onkeydown={(e) => e.key === 'Enter' && onSelectCommit(commit.sha)}
        >
          <div class="commit-bullet">
            <div class="bullet-dot"></div>
          </div>

          <div class="flex-1">
            <div class="flex items-start justify-between gap-4 mb-1">
              <span class="commit-msg font-semibold text-sm text-main hover:text-emerald-400 transition-colors">
                {commit.message.split('\n')[0]}
              </span>

              <!-- SHA Copy Badge -->
              <button
                class="sha-badge font-mono text-xs flex items-center gap-1.5"
                onclick={(e) => copyHash(e, commit.sha)}
                title="Copy Full SHA"
              >
                <span>{commit.sha.slice(0, 7)}</span>
                {#if copiedSHA === commit.sha}
                  <Check class="w-3 h-3 text-emerald-400" />
                {:else}
                  <Copy class="w-3 h-3 text-dim" />
                {/if}
              </button>
            </div>

            <div class="flex items-center gap-3 text-xs text-dim">
              <span class="flex items-center gap-1 text-muted">
                <User class="w-3 h-3" />
                <span>{commit.author?.name || 'Developer'}</span>
              </span>
              <span>•</span>
              <span class="flex items-center gap-1">
                <Calendar class="w-3 h-3" />
                <span>{new Date(commit.timestamp).toLocaleString()}</span>
              </span>
            </div>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .commit-list {
    padding: 0;
    overflow: hidden;
  }

  .commit-header {
    padding: 0.75rem 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .commit-row {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem 1.25rem;
    border-bottom: 1px solid rgba(36, 48, 72, 0.4);
    transition: background 0.15s ease;
    position: relative;
  }

  .commit-row:last-child {
    border-bottom: none;
  }

  .commit-row:hover {
    background: rgba(30, 41, 59, 0.5);
  }

  .commit-bullet {
    margin-top: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
  }

  .bullet-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--accent-emerald);
    box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
  }

  .sha-badge {
    padding: 0.2rem 0.5rem;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
    color: var(--text-muted);
    transition: all 0.15s ease;
  }

  .sha-badge:hover {
    background: var(--bg-card-hover);
    border-color: #3b4d6b;
    color: var(--text-main);
  }
</style>
