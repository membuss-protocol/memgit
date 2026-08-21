<script>
  import { CircleDot, CheckCircle2, MessageSquare, Plus, Send, User, Tag, ArrowLeft, X } from 'lucide-svelte';
  import { api } from '../api/client.js';

  let { repoName } = $props();

  let issues = $state([]);
  let activeTab = $state('open'); // 'open' | 'closed'
  let selectedIssue = $state(null);
  let showNewIssueModal = $state(false);

  // New Issue Form
  let newTitle = $state('');
  let newBody = $state('');
  let newAuthor = $state('Developer');

  // Comment Form
  let commentBody = $state('');
  let commentAuthor = $state('Developer');

  let loading = $state(false);

  $effect(() => {
    loadIssues();
  });

  async function loadIssues() {
    try {
      issues = await api.getIssues(repoName);
    } catch (e) {
      console.error(e);
    }
  }

  let filteredIssues = $derived(
    issues.filter((i) => i.state === activeTab)
  );

  async function handleCreateIssue(e) {
    e.preventDefault();
    if (!newTitle.trim()) return;
    loading = true;
    try {
      await api.createIssue(repoName, {
        title: newTitle.trim(),
        body: newBody.trim(),
        author: newAuthor.trim() || 'Developer',
        labels: ['proposal'],
      });
      newTitle = '';
      newBody = '';
      showNewIssueModal = false;
      await loadIssues();
    } catch (err) {
      alert(err.message);
    } finally {
      loading = false;
    }
  }

  async function handleAddComment(e) {
    e.preventDefault();
    if (!commentBody.trim() || !selectedIssue) return;
    try {
      await api.addIssueComment(repoName, selectedIssue.id, {
        author: commentAuthor.trim() || 'Developer',
        body: commentBody.trim(),
      });
      commentBody = '';
      await loadIssues();
      selectedIssue = issues.find((i) => i.id === selectedIssue.id);
    } catch (err) {
      alert(err.message);
    }
  }

  async function handleToggleState() {
    if (!selectedIssue) return;
    const newState = selectedIssue.state === 'open' ? 'closed' : 'open';
    try {
      await api.updateIssueState(repoName, selectedIssue.id, newState);
      await loadIssues();
      selectedIssue = issues.find((i) => i.id === selectedIssue.id);
    } catch (err) {
      alert(err.message);
    }
  }
</script>

<div class="issues-page">
  {#if selectedIssue}
    <!-- Issue Detail View -->
    <div class="issue-detail">
      <div class="flex items-center justify-between gap-4 mb-4 pb-4 border-b border-slate-800">
        <button class="btn btn-ghost btn-sm flex items-center gap-1.5" onclick={() => selectedIssue = null}>
          <ArrowLeft class="w-4 h-4 text-emerald-400" />
          <span>Back to Issues</span>
        </button>

        <button class="btn btn-secondary btn-sm" onclick={handleToggleState}>
          {#if selectedIssue.state === 'open'}
            <CheckCircle2 class="w-4 h-4 text-purple-400" />
            <span>Close Issue</span>
          {:else}
            <CircleDot class="w-4 h-4 text-emerald-400" />
            <span>Reopen Issue</span>
          {/if}
        </button>
      </div>

      <div class="mb-6">
        <div class="flex items-center gap-2 mb-2">
          <h2 class="text-2xl font-bold text-main">{selectedIssue.title}</h2>
          <span class="text-xl text-dim font-mono">#{selectedIssue.id}</span>
        </div>

        <div class="flex items-center gap-3 text-xs text-muted">
          <span class="badge {selectedIssue.state === 'open' ? 'badge-emerald' : 'badge-purple'}">
            {selectedIssue.state === 'open' ? 'Open' : 'Closed'}
          </span>
          <span>Opened by <strong>{selectedIssue.author}</strong> on {new Date(selectedIssue.created_at).toLocaleDateString()}</span>
        </div>
      </div>

      <!-- Main Body Comment -->
      <div class="card card-elevated p-5 mb-6">
        <div class="text-sm text-main whitespace-pre-wrap leading-relaxed">
          {selectedIssue.body || 'No description provided.'}
        </div>
      </div>

      <!-- Comments Stream -->
      <div class="space-y-4 mb-8">
        {#each selectedIssue.comments || [] as comment}
          <div class="card p-4">
            <div class="flex items-center justify-between text-xs text-dim mb-2 pb-2 border-b border-slate-800">
              <span class="text-main font-semibold flex items-center gap-1.5">
                <User class="w-3.5 h-3.5 text-cyan-400" />
                <span>{comment.author}</span>
              </span>
              <span>{new Date(comment.created_at).toLocaleString()}</span>
            </div>
            <p class="text-sm text-muted whitespace-pre-wrap">{comment.body}</p>
          </div>
        {/each}
      </div>

      <!-- Add Comment Box -->
      <form onsubmit={handleAddComment} class="card card-elevated p-5">
        <h4 class="text-xs font-semibold text-main mb-2 font-mono">Add a Comment</h4>
        <textarea
          bind:value={commentBody}
          rows="3"
          placeholder="Leave a comment or discussion point..."
          class="input-field mb-3"
          required
        ></textarea>
        <div class="flex items-center justify-between">
          <input
            type="text"
            bind:value={commentAuthor}
            placeholder="Your name"
            class="input-field max-w-xs text-xs font-mono"
          />
          <button type="submit" class="btn btn-primary btn-sm">
            <Send class="w-3.5 h-3.5" />
            <span>Comment</span>
          </button>
        </div>
      </form>
    </div>
  {:else}
    <!-- Issues Listing View -->
    <div class="card card-elevated">
      <!-- Action Bar -->
      <div class="flex items-center justify-between p-4 border-b border-slate-800 bg-slate-900/60">
        <div class="flex items-center gap-2">
          <button
            class="btn btn-sm {activeTab === 'open' ? 'btn-primary' : 'btn-ghost'}"
            onclick={() => activeTab = 'open'}
          >
            <CircleDot class="w-3.5 h-3.5" />
            <span>Open ({issues.filter((i) => i.state === 'open').length})</span>
          </button>
          <button
            class="btn btn-sm {activeTab === 'closed' ? 'btn-primary' : 'btn-ghost'}"
            onclick={() => activeTab = 'closed'}
          >
            <CheckCircle2 class="w-3.5 h-3.5" />
            <span>Closed ({issues.filter((i) => i.state === 'closed').length})</span>
          </button>
        </div>

        <button class="btn btn-primary btn-sm" onclick={() => showNewIssueModal = true}>
          <Plus class="w-3.5 h-3.5" />
          <span>New Issue</span>
        </button>
      </div>

      <!-- Issue Items -->
      <div class="divide-y divide-slate-800">
        {#if filteredIssues.length === 0}
          <div class="p-12 text-center text-muted text-sm">
            No {activeTab} issues in this repository.
          </div>
        {:else}
          {#each filteredIssues as issue}
            <div
              class="p-4 hover:bg-slate-800/40 transition-colors cursor-pointer flex items-start justify-between gap-4"
              onclick={() => selectedIssue = issue}
              role="button"
              tabindex="0"
              onkeydown={(e) => e.key === 'Enter' && (selectedIssue = issue)}
            >
              <div class="flex items-start gap-3">
                {#if issue.state === 'open'}
                  <CircleDot class="w-4 h-4 text-emerald-400 mt-1 shrink-0" />
                {:else}
                  <CheckCircle2 class="w-4 h-4 text-purple-400 mt-1 shrink-0" />
                {/if}
                <div>
                  <h4 class="text-sm font-semibold text-main hover:text-emerald-400 transition-colors mb-1">
                    {issue.title}
                  </h4>
                  <div class="flex items-center gap-2 text-xs text-dim">
                    <span>#{issue.id}</span>
                    <span>•</span>
                    <span>Opened by {issue.author} on {new Date(issue.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>

              {#if issue.comments?.length > 0}
                <div class="flex items-center gap-1.5 text-xs text-muted font-mono">
                  <MessageSquare class="w-3.5 h-3.5" />
                  <span>{issue.comments.length}</span>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
    </div>
  {/if}

  <!-- New Issue Modal -->
  {#if showNewIssueModal}
    <div
      class="modal-overlay"
      onclick={() => showNewIssueModal = false}
      onkeydown={(e) => e.key === 'Escape' && (showNewIssueModal = false)}
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
          <h3 class="text-base font-bold text-main font-mono">Create New Issue</h3>
          <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={() => showNewIssueModal = false} aria-label="Close modal">
            <X class="w-4 h-4" />
          </button>
        </div>
        <form onsubmit={handleCreateIssue} class="p-6">
          <div class="mb-4">
            <label for="issue-title" class="block text-xs font-semibold text-main mb-1.5 font-mono">Title *</label>
            <input
              id="issue-title"
              type="text"
              bind:value={newTitle}
              placeholder="Issue title"
              class="input-field"
              required
            />
          </div>
          <div class="mb-4">
            <label for="issue-body" class="block text-xs font-semibold text-main mb-1.5">Description</label>
            <textarea
              id="issue-body"
              bind:value={newBody}
              rows="4"
              placeholder="Describe the bug or feature proposal..."
              class="input-field"
            ></textarea>
          </div>
          <div class="mb-6">
            <label for="issue-author" class="block text-xs font-semibold text-main mb-1.5 font-mono">Author Name</label>
            <input
              id="issue-author"
              type="text"
              bind:value={newAuthor}
              placeholder="Your name"
              class="input-field font-mono"
            />
          </div>
          <div class="flex justify-end gap-3 pt-4 border-t border-slate-800">
            <button type="button" class="btn btn-secondary text-xs" onclick={() => showNewIssueModal = false}>Cancel</button>
            <button type="submit" class="btn btn-primary text-xs" disabled={loading}>Submit Issue</button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
