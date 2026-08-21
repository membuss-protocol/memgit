<script>
  import { onMount } from 'svelte';
  import { Search, Compass, Star, GitFork, Shield, Radio, Server, Terminal, Filter, Flame, Globe } from 'lucide-svelte';
  import { api } from '../api/client.js';
  import RepoCard from '../components/RepoCard.svelte';
  import CloneModal from '../components/CloneModal.svelte';

  let { initialFilter = 'all', initialQuery = '', onSelectRepo, onNewRepo } = $props();

  let repos = $state([]);
  let loading = $state(true);
  let error = $state('');
  let activeFilter = $state('all');
  let searchQuery = $state('');
  let activeTopic = $state('');
  let selectedCloneRepo = $state(null);

  $effect(() => {
    activeFilter = initialFilter || 'all';
    searchQuery = initialQuery || '';
  });

  let dynamicTopics = $derived.by(() => {
    const set = new Set();
    repos.forEach(r => {
      if (Array.isArray(r.topics)) {
        r.topics.forEach(t => {
          if (t && t.trim()) set.add(t.trim());
        });
      }
    });
    return Array.from(set);
  });

  async function loadRepos() {
    loading = true;
    error = '';
    try {
      const data = await api.getExploreRepos(activeFilter, searchQuery);
      repos = data;
    } catch (err) {
      error = err.message || 'Failed to load swarm repositories';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadRepos();
  });

  function handleFilterChange(filter) {
    activeFilter = filter;
    loadRepos();
  }

  function handleSearchKeydown(e) {
    if (e.key === 'Enter') {
      loadRepos();
    }
  }

  let filteredRepos = $derived.by(() => {
    if (!activeTopic) return repos;
    return repos.filter(r => r.topics?.includes(activeTopic));
  });

  async function handleStarToggle(repo, e) {
    e.stopPropagation();
    try {
      const res = await api.starRepo(repo.name);
      repo.star_count = res.star_count;
      repo.is_starred = res.is_starred;
    } catch (err) {
      console.error('Star toggle failed', err);
    }
  }

  async function handleFork(repo, e) {
    e.stopPropagation();
    try {
      const forked = await api.forkRepo(repo.full_name || repo.name);
      if (onSelectRepo) onSelectRepo(forked.name);
    } catch (err) {
      alert('Fork failed: ' + err.message);
    }
  }
</script>

<div class="max-w-7xl mx-auto px-4 lg:px-8 py-8">
  <!-- Header Banner -->
  <div class="mb-8 p-6 lg:p-8 rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900/90 to-slate-950 border border-slate-800/80 shadow-2xl relative overflow-hidden">
    <div class="absolute -right-10 -bottom-10 w-72 h-72 bg-emerald-500/5 rounded-full blur-3xl pointer-events-none"></div>
    <div class="relative z-10">
      <div class="flex items-center gap-2 text-xs font-mono text-cyan-400 mb-2">
        <Compass class="w-4 h-4" />
        <span>Membuss Decentralized Network Catalog</span>
      </div>
      <h1 class="text-2xl lg:text-3xl font-extrabold text-main tracking-tight mb-2">
        Explore Global Repositories
      </h1>
      <p class="text-xs lg:text-sm text-muted max-w-2xl leading-relaxed">
        Discover open-source code repositories distributed across the Membuss peer-to-peer swarm. Every project is chunked into Merkle DAGs, verified with Ed25519 MemNS signatures, and replicated with 10+4 Reed-Solomon parity.
      </p>
    </div>
  </div>

  <!-- Search & Filter Controls -->
  <div class="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-4 mb-6">
    <!-- Filter Tabs -->
    <div class="flex items-center gap-1.5 p-1 bg-slate-900/80 rounded-xl border border-slate-800 self-start">
      <button
        class="filter-btn {activeFilter === 'all' ? 'filter-btn-active' : ''}"
        onclick={() => handleFilterChange('all')}
      >
        <Globe class="w-3.5 h-3.5 inline mr-1" />
        All Swarm ({repos.length})
      </button>
      <button
        class="filter-btn {activeFilter === 'local' ? 'filter-btn-active' : ''}"
        onclick={() => handleFilterChange('local')}
      >
        <Server class="w-3.5 h-3.5 inline mr-1" />
        Local Node
      </button>
      <button
        class="filter-btn {activeFilter === 'starred' ? 'filter-btn-active' : ''}"
        onclick={() => handleFilterChange('starred')}
      >
        <Star class="w-3.5 h-3.5 inline mr-1 text-amber-400" />
        Starred
      </button>
    </div>

    <!-- Search Input -->
    <div class="relative flex-1 max-w-md">
      <Search class="w-4 h-4 text-muted absolute left-3 top-1/2 -translate-y-1/2" />
      <input
        type="text"
        bind:value={searchQuery}
        onkeydown={handleSearchKeydown}
        placeholder="Filter by name, owner, description..."
        class="input-field pl-9 pr-3 py-2 text-xs w-full font-sans bg-slate-950/70"
      />
    </div>
  </div>

  <!-- Topic Filter Chips (if any real topics exist) -->
  {#if dynamicTopics.length > 0}
    <div class="flex items-center gap-1.5 overflow-x-auto pb-2 mb-6 scrollbar-none">
      <span class="text-[11px] font-mono text-muted uppercase tracking-wider mr-1 flex items-center gap-1">
        <Filter class="w-3 h-3" /> Topics:
      </span>
      <button
        class="px-2.5 py-1 rounded-lg text-xs font-mono transition-all {!activeTopic ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 font-semibold' : 'bg-slate-900 text-muted border border-slate-800 hover:text-main'}"
        onclick={() => activeTopic = ''}
      >
        all
      </button>
      {#each dynamicTopics as topic}
        <button
          class="px-2.5 py-1 rounded-lg text-xs font-mono transition-all {activeTopic === topic ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 font-semibold' : 'bg-slate-900 text-muted border border-slate-800 hover:text-main'}"
          onclick={() => activeTopic = topic}
        >
          #{topic}
        </button>
      {/each}
    </div>
  {/if}

  <!-- Repository Grid -->
  {#if loading}
    <div class="p-12 text-center text-muted font-mono text-xs">
      <div class="animate-spin w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full mx-auto mb-3"></div>
      Discovering repositories across Membuss DHT swarm...
    </div>
  {:else if error}
    <div class="p-6 bg-red-950/30 border border-red-800/40 rounded-xl text-xs text-red-400 text-center">
      {error}
    </div>
  {:else if filteredRepos.length === 0}
    <div class="p-12 text-center bg-slate-900/30 border border-slate-800 rounded-xl text-muted text-xs font-mono">
      No repositories match your criteria. Try adjusting your search query or filter.
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      {#each filteredRepos as repo (repo.name || repo.full_name)}
        <div class="card p-5 flex flex-col justify-between group hover:border-slate-700 transition-all duration-180">
          <div>
            <!-- Header: Owner / Name & Badges -->
            <div class="flex items-start justify-between gap-3 mb-2.5">
              <button
                class="text-left font-mono font-bold text-sm text-main hover:text-emerald-400 transition-colors truncate"
                onclick={() => onSelectRepo(repo.name)}
              >
                <span class="text-muted font-normal">@{repo.owner || 'dev'}/</span>{repo.name}
              </button>
              <div class="flex items-center gap-1.5 shrink-0">
                {#if repo.is_local}
                  <span class="badge badge-emerald text-[10px] py-0.5 px-1.5">Local</span>
                {:else}
                  <span class="badge badge-cyan text-[10px] py-0.5 px-1.5">Swarm</span>
                {/if}
              </div>
            </div>

            <!-- Description -->
            <p class="text-xs text-muted leading-relaxed line-clamp-2 mb-4">
              {repo.description || 'No description provided.'}
            </p>

            <!-- Topics -->
            {#if repo.topics?.length > 0}
              <div class="flex flex-wrap gap-1.5 mb-4">
                {#each repo.topics.slice(0, 3) as topic}
                  <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-900 text-slate-400 border border-slate-800">
                    #{topic}
                  </span>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Footer Actions & Metadata -->
          <div class="pt-4 border-t border-slate-800/80 flex items-center justify-between gap-2 text-xs">
            <div class="flex items-center gap-3 text-muted font-mono text-[11px]">
              <button
                class="flex items-center gap-1 hover:text-amber-400 transition-colors"
                onclick={(e) => handleStarToggle(repo, e)}
                title="Star repository"
              >
                <Star class="w-3.5 h-3.5 {repo.is_starred ? 'fill-amber-400 text-amber-400' : ''}" />
                <span>{repo.star_count || 0}</span>
              </button>
              <button
                class="flex items-center gap-1 hover:text-cyan-400 transition-colors"
                onclick={(e) => handleFork(repo, e)}
                title="Fork repository"
              >
                <GitFork class="w-3.5 h-3.5" />
                <span>{repo.fork_count || 0}</span>
              </button>
            </div>

            <div class="flex items-center gap-2">
              <button
                class="btn btn-secondary btn-sm py-1 px-2.5 text-[11px]"
                onclick={(e) => { e.stopPropagation(); selectedCloneRepo = repo; }}
              >
                Clone
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if selectedCloneRepo}
  <CloneModal repo={selectedCloneRepo} onClose={() => selectedCloneRepo = null} />
{/if}

<style>
  .filter-btn {
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-text-muted);
    border-radius: 0.5rem;
    transition: all 150ms ease;
  }
  .filter-btn:hover {
    color: var(--color-text-main);
  }
  .filter-btn-active {
    background-color: var(--color-surface-hover);
    color: var(--color-text-main);
    font-weight: 600;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  }
</style>
