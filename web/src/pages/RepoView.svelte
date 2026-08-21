<script>
  import { onMount } from 'svelte';
  import {
    BookOpen, Star, GitBranch, GitCommit, CircleDot, GitPullRequest,
    Tag, Radio, Terminal, RefreshCw, ChevronDown, Check, Shield, Key, GitFork
  } from 'lucide-svelte';
  import { api } from '../api/client.js';
  import FileTree from '../components/FileTree.svelte';
  import CodeViewer from '../components/CodeViewer.svelte';
  import MarkdownViewer from '../components/MarkdownViewer.svelte';
  import CommitList from '../components/CommitList.svelte';
  import DiffViewer from '../components/DiffViewer.svelte';
  import CloneModal from '../components/CloneModal.svelte';
  import SwarmRadar from '../components/SwarmRadar.svelte';
  import IssuesPage from './IssuesPage.svelte';
  import PullRequestsPage from './PullRequestsPage.svelte';
  import ReleasesPage from './ReleasesPage.svelte';

  let { repoName, onGoBack, onNavigate } = $props();

  let repo = $state(null);
  let activeTab = $state('code'); // 'code' | 'commits' | 'issues' | 'pulls' | 'releases' | 'swarm'
  let branches = $state([]);
  let currentBranch = $state('main');
  let currentPath = $state('');
  let tree = $state([]);
  let currentBlob = $state(null);
  let currentBlobPath = $state('');
  let commits = $state([]);
  let selectedCommitSHA = $state(null);
  let commitDiff = $state(null);
  let swarm = $state(null);
  let readmeContent = $state('');

  let showCloneModal = $state(false);
  let syncing = $state(false);
  let forking = $state(false);

  $effect(() => {
    if (repoName) {
      loadAll();
    }
  });

  async function loadAll() {
    try {
      repo = await api.getRepo(repoName);
      branches = await api.getBranches(repoName).catch(() => []);
      if (branches.length > 0 && !currentBranch) {
        currentBranch = branches.find((b) => b.is_default)?.name || branches[0].name;
      }
      await loadTree();
      await loadCommits();
      await loadSwarm();
    } catch (e) {
      console.error(e);
    }
  }

  async function loadTree() {
    try {
      tree = await api.getTree(repoName, currentBranch, currentPath);
      // Auto-load README if at root
      if (!currentPath) {
        const readme = tree.find((t) => t.name.toLowerCase() === 'readme.md');
        if (readme) {
          const blob = await api.getBlob(repoName, currentBranch, readme.path);
          readmeContent = blob.content;
        } else {
          readmeContent = '';
        }
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function loadCommits() {
    try {
      commits = await api.getCommits(repoName, currentBranch, 30);
    } catch (e) {
      console.error(e);
    }
  }

  async function loadSwarm() {
    try {
      swarm = await api.getSwarmStatus(repoName);
    } catch (e) {
      console.error(e);
    }
  }

  async function handleSelectFile(path) {
    try {
      currentBlobPath = path;
      currentBlob = await api.getBlob(repoName, currentBranch, path);
    } catch (e) {
      alert(e.message);
    }
  }

  async function handleSelectCommit(sha) {
    try {
      selectedCommitSHA = sha;
      commitDiff = await api.getCommit(repoName, sha);
    } catch (e) {
      alert(e.message);
    }
  }

  async function handleSync() {
    syncing = true;
    try {
      await api.syncRepo(repoName);
      await loadAll();
    } catch (err) {
      alert(err.message);
    } finally {
      syncing = false;
    }
  }

  async function handleStar() {
    try {
      const res = await api.starRepo(repoName);
      if (repo) {
        repo.star_count = res.star_count;
        repo.is_starred = res.is_starred;
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function handleFork() {
    forking = true;
    try {
      const forked = await api.forkRepo(repo?.full_name || repoName);
      alert(`Forked successfully to @${forked.owner}/${forked.name}!`);
      if (onGoBack) onGoBack();
    } catch (e) {
      alert('Fork failed: ' + e.message);
    } finally {
      forking = false;
    }
  }
</script>

{#if !repo}
  <div class="card p-12 text-center text-muted font-mono">
    Loading repository {repoName}...
  </div>
{:else}
  <div class="repo-view max-w-7xl mx-auto px-4 lg:px-8 py-8">
    <!-- Header -->
    <div class="repo-header card p-6 mb-6">
      <div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-4">
        <!-- Title & Breadcrumbs -->
        <div class="flex items-center gap-2.5 flex-wrap">
          <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={onGoBack}>
            ← Explore
          </button>
          <span class="text-muted">/</span>
          {#if repo.owner}
            <button
              class="font-mono text-sm text-cyan-400 hover:underline"
              onclick={() => onNavigate && onNavigate('user', { username: repo.owner })}
            >
              @{repo.owner}
            </button>
            <span class="text-muted font-mono">/</span>
          {/if}
          <h1 class="text-lg font-bold font-mono text-main">{repo.name}</h1>
          <span class="badge {repo.is_private ? 'badge-amber' : 'badge-slate'} text-[10px]">
            {repo.is_private ? 'Private' : 'Public'}
          </span>
          {#if repo.forked_from}
            <span class="text-[11px] font-mono text-muted flex items-center gap-1 ml-1">
              <GitFork class="w-3 h-3 text-cyan-400" /> forked from {repo.forked_from}
            </span>
          {/if}
        </div>

        <!-- Top Actions -->
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Star -->
          <button class="btn btn-secondary btn-sm" onclick={handleStar}>
            <Star class="w-3.5 h-3.5 text-amber-400 {repo.is_starred ? 'fill-amber-400' : ''}" />
            <span class="font-mono text-xs">Star ({repo.star_count || 0})</span>
          </button>

          <!-- Fork -->
          <button class="btn btn-secondary btn-sm" onclick={handleFork} disabled={forking} title="Fork to your local node">
            <GitFork class="w-3.5 h-3.5 text-cyan-400" />
            <span class="text-xs font-mono">Fork ({repo.fork_count || 0})</span>
          </button>

          <!-- Sync to Membuss -->
          {#if repo.is_local}
            <button class="btn btn-secondary btn-sm" onclick={handleSync} disabled={syncing} title="Force Merkle DAG Snapshot & Sign MemNS">
              <RefreshCw class="w-3.5 h-3.5 text-emerald-400 {syncing ? 'animate-spin' : ''}" />
              <span class="text-xs">Sync Membuss</span>
            </button>
          {/if}

          <!-- Clone Button -->
          <button class="btn btn-primary btn-sm" onclick={() => showCloneModal = true}>
            <Terminal class="w-3.5 h-3.5" />
            <span>Clone</span>
          </button>
        </div>
      </div>

      <!-- Description & MemNS Bar -->
      <p class="text-xs text-muted mb-4 leading-relaxed">
        {repo.description || 'Decentralized content-addressed repository running on Membuss.'}
      </p>

      <div class="flex flex-wrap items-center gap-3 pt-3 border-t border-slate-800 text-xs font-mono">
        <div class="flex items-center gap-1.5 text-cyan-300">
          <Key class="w-3.5 h-3.5 text-cyan-400" />
          <span>MemNS: /memns/{repo.memns_name}</span>
        </div>
        {#if repo.latest_mid}
          <span class="text-slate-600">•</span>
          <div class="flex items-center gap-1.5 text-emerald-300">
            <Shield class="w-3.5 h-3.5 text-emerald-400" />
            <span>Latest MID: {repo.latest_mid}</span>
          </div>
        {/if}
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="tabs-bar flex items-center gap-1 mb-6 border-b border-slate-800 pb-px overflow-x-auto">
      <button
        class="tab-item {activeTab === 'code' ? 'tab-active' : ''}"
        onclick={() => { activeTab = 'code'; currentBlob = null; }}
      >
        <BookOpen class="w-4 h-4" />
        <span>Code</span>
      </button>

      <button
        class="tab-item {activeTab === 'commits' ? 'tab-active' : ''}"
        onclick={() => { activeTab = 'commits'; selectedCommitSHA = null; }}
      >
        <GitCommit class="w-4 h-4" />
        <span>Commits ({commits.length})</span>
      </button>

      <button
        class="tab-item {activeTab === 'issues' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'issues'}
      >
        <CircleDot class="w-4 h-4" />
        <span>Issues</span>
      </button>

      <button
        class="tab-item {activeTab === 'pulls' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'pulls'}
      >
        <GitPullRequest class="w-4 h-4" />
        <span>Pull Requests</span>
      </button>

      <button
        class="tab-item {activeTab === 'releases' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'releases'}
      >
        <Tag class="w-4 h-4" />
        <span>Releases</span>
      </button>

      <button
        class="tab-item {activeTab === 'swarm' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'swarm'}
      >
        <Radio class="w-4 h-4" />
        <span>Swarm Telemetry</span>
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      {#if activeTab === 'code'}
        {#if currentBlob}
          <!-- Code Viewer Mode -->
          <div class="mb-4">
            <button
              class="btn btn-ghost btn-sm text-xs font-mono text-muted hover:text-main"
              onclick={() => currentBlob = null}
            >
              ← Back to {repo.name}/{currentPath}
            </button>
          </div>
          <CodeViewer blob={currentBlob} repoName={repo.name} refName={currentBranch} />
        {:else}
          <!-- Tree Mode -->
          <FileTree
            {tree}
            {branches}
            {currentBranch}
            {currentPath}
            onSelectFile={handleSelectFile}
            onNavigateDir={(dir) => { currentPath = dir; loadTree(); }}
            onChangeBranch={(b) => { currentBranch = b; loadTree(); loadCommits(); }}
          />

          <!-- README Section -->
          {#if readmeContent}
            <div class="mt-8">
              <MarkdownViewer content={readmeContent} filename="README.md" />
            </div>
          {/if}
        {/if}
      {:else if activeTab === 'commits'}
        {#if selectedCommitSHA && commitDiff}
          <div class="mb-4">
            <button
              class="btn btn-ghost btn-sm text-xs font-mono text-muted hover:text-main"
              onclick={() => { selectedCommitSHA = null; commitDiff = null; }}
            >
              ← Back to Commit History
            </button>
          </div>
          <DiffViewer diff={commitDiff} />
        {:else}
          <CommitList {commits} onSelectCommit={handleSelectCommit} />
        {/if}
      {:else if activeTab === 'issues'}
        <IssuesPage repoName={repo.name} />
      {:else if activeTab === 'pulls'}
        <PullRequestsPage repoName={repo.name} />
      {:else if activeTab === 'releases'}
        <ReleasesPage repoName={repo.name} />
      {:else if activeTab === 'swarm'}
        <SwarmRadar />
      {/if}
    </div>
  </div>
{/if}

{#if showCloneModal && repo}
  <CloneModal {repo} onClose={() => showCloneModal = false} />
{/if}

<style>
  .tab-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--color-text-muted);
    border-bottom: 2px solid transparent;
    transition: all 150ms ease;
    white-space: nowrap;
  }
  .tab-item:hover {
    color: var(--color-text-main);
  }
  .tab-active {
    color: var(--color-text-main);
    font-weight: 600;
    border-bottom-color: var(--color-brand);
  }
</style>
