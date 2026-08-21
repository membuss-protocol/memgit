<script>
  import { onMount } from 'svelte';
  import { Plus, Search, Terminal, Radio, Shield, Star, GitFork, Activity, Globe, Compass, ChevronRight, CheckCircle2, Server } from 'lucide-svelte';
  import { api } from '../api/client.js';
  import SwarmRadar from '../components/SwarmRadar.svelte';

  let { currentUser, onSelectRepo, onNewRepo, onNavigate } = $props();

  let repos = $state([]);
  let activities = $state([]);
  let status = $state(null);
  let loading = $state(true);
  let repoSearch = $state('');

  async function loadData() {
    loading = true;
    try {
      const [r, act, s] = await Promise.all([
        api.getRepos().catch(() => []),
        api.getActivityFeed(20).catch(() => []),
        api.getSystemStatus().catch(() => null),
      ]);
      repos = r;
      activities = act;
      status = s;
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadData();
  });

  let filteredRepos = $derived.by(() => {
    if (!repoSearch.trim()) return repos;
    const q = repoSearch.toLowerCase();
    return repos.filter(r => r.name.toLowerCase().includes(q) || (r.description && r.description.toLowerCase().includes(q)));
  });

  function getAvatarColor(name) {
    const colors = ['bg-emerald-600', 'bg-cyan-600', 'bg-purple-600', 'bg-blue-600', 'bg-amber-600'];
    let hash = 0;
    for (let i = 0; i < (name || '').length; i++) hash += name.charCodeAt(i);
    return colors[Math.abs(hash) % colors.length];
  }

  function getActivityIcon(type) {
    switch (type) {
      case 'star': return '⭐';
      case 'fork': return '🍴';
      case 'pr': return '🔀';
      case 'pr_merged': return '🟣';
      case 'issue': return '⚠️';
      case 'release': return '🏷️';
      default: return '📦';
    }
  }
</script>

<div class="max-w-7xl mx-auto px-4 lg:px-8 py-8">
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
    <!-- Left Column: User Profile & Local Repositories (4 cols) -->
    <div class="lg:col-span-4 space-y-6">
      <!-- User Profile Summary Card -->
      <div class="card p-5">
        <div class="flex items-center gap-3.5 mb-4">
          <div class="w-12 h-12 rounded-xl {getAvatarColor(currentUser?.username)} text-white text-xl font-bold flex items-center justify-center font-mono uppercase shadow-md">
            {(currentUser?.username || 'D')[0]}
          </div>
          <div class="min-w-0 flex-1">
            <h3 class="text-sm font-bold text-main truncate">{currentUser?.display_name || 'Membuss Developer'}</h3>
            <p class="text-xs font-mono text-muted truncate">@{currentUser?.username || 'developer'}</p>
          </div>
        </div>

        {#if currentUser?.bio}
          <p class="text-xs text-muted leading-relaxed mb-4 p-2.5 bg-slate-900/60 rounded-lg border border-slate-800">
            {currentUser.bio}
          </p>
        {/if}

        <div class="grid grid-cols-2 gap-2 text-center text-xs font-mono pt-3 border-t border-slate-800">
          <div class="p-2 bg-slate-900 rounded-lg border border-slate-800">
            <div class="text-main font-bold">{repos.length}</div>
            <div class="text-[10px] text-muted uppercase">My Repos</div>
          </div>
          <div class="p-2 bg-slate-900 rounded-lg border border-slate-800">
            <div class="text-emerald-400 font-bold">10+4 RS</div>
            <div class="text-[10px] text-muted uppercase">Parity</div>
          </div>
        </div>
      </div>

      <!-- My Repositories List -->
      <div class="card p-5">
        <div class="flex items-center justify-between gap-2 mb-3">
          <h4 class="text-xs font-bold text-main font-mono uppercase tracking-wider flex items-center gap-1.5">
            <Server class="w-3.5 h-3.5 text-emerald-400" />
            My Repositories ({repos.length})
          </h4>
          <button class="btn btn-primary btn-sm py-1 px-2 text-[11px] flex items-center gap-1" onclick={onNewRepo}>
            <Plus class="w-3 h-3" /> New
          </button>
        </div>

        <!-- Search repos input -->
        <div class="relative mb-3">
          <Search class="w-3.5 h-3.5 text-muted absolute left-2.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            bind:value={repoSearch}
            placeholder="Filter local repos..."
            class="input-field pl-8 pr-2.5 py-1 text-xs w-full font-mono bg-slate-950/70"
          />
        </div>

        <!-- Repo list -->
        {#if loading}
          <div class="py-6 text-center text-xs text-muted font-mono">Loading...</div>
        {:else if filteredRepos.length === 0}
          <div class="p-4 text-center text-xs text-muted bg-slate-900/40 rounded-lg border border-slate-800 font-mono">
            No repositories found.
          </div>
        {:else}
          <div class="space-y-2 max-h-96 overflow-y-auto pr-1">
            {#each filteredRepos as repo}
              <button
                class="w-full text-left p-3 rounded-lg bg-slate-900/50 hover:bg-slate-900 border border-slate-800/80 hover:border-slate-700 transition-all flex items-center justify-between group"
                onclick={() => onSelectRepo(repo.name)}
              >
                <div class="min-w-0 pr-2">
                  <div class="font-mono text-xs font-bold text-main group-hover:text-emerald-400 transition-colors truncate">
                    {repo.name}
                  </div>
                  <div class="text-[11px] text-muted truncate">
                    {repo.description || 'No description'}
                  </div>
                </div>
                <ChevronRight class="w-3.5 h-3.5 text-muted group-hover:text-main group-hover:translate-x-0.5 transition-all shrink-0" />
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Quick Git Clone Helper -->
      <div class="p-4 bg-slate-900/40 border border-slate-800 rounded-xl text-xs space-y-2 font-mono text-muted">
        <div class="text-[11px] font-bold text-main flex items-center gap-1.5 uppercase tracking-wider">
          <Terminal class="w-3.5 h-3.5 text-cyan-400" /> Push-to-Create
        </div>
        <p class="text-[11px] text-slate-400 leading-relaxed font-sans">
          Push to any URL to initialize a repository automatically:
        </p>
        <div class="p-2 bg-slate-950 rounded border border-slate-800 text-[10px] text-emerald-400 select-all overflow-x-auto">
          git remote add origin http://localhost:8500/git/my-app.git<br>
          git push -u origin main
        </div>
      </div>
    </div>

    <!-- Center/Right Column: Swarm Feed & Global Highlights (8 cols) -->
    <div class="lg:col-span-8 space-y-6">
      <!-- Global Explore Banner -->
      <div class="p-6 rounded-2xl bg-gradient-to-r from-emerald-950/40 via-slate-900 to-cyan-950/30 border border-slate-800 shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2 text-xs font-mono text-emerald-400 mb-1">
            <Radio class="w-3.5 h-3.5 animate-pulse" />
            <span>Membuss P2P Network Live</span>
          </div>
          <h2 class="text-lg font-bold text-main tracking-tight">Decentralized Collaboration Swarm</h2>
          <p class="text-xs text-muted max-w-lg mt-0.5">
            Discover peer repositories, browse code diffs, review pull requests, and clone directly over the public gateway CDN.
          </p>
        </div>
        <button
          class="btn btn-primary text-xs shrink-0 flex items-center gap-1.5"
          onclick={() => onNavigate('explore')}
        >
          <Compass class="w-3.5 h-3.5" />
          <span>Explore All Repos</span>
        </button>
      </div>

      <!-- Global Activity Feed -->
      <div class="card p-6">
        <div class="flex items-center justify-between gap-2 mb-5 pb-3 border-b border-slate-800">
          <div class="flex items-center gap-2 font-mono text-xs font-bold text-main uppercase tracking-wider">
            <Activity class="w-4 h-4 text-cyan-400" />
            <span>Global Swarm Activity</span>
          </div>
          <span class="text-[11px] text-muted font-mono">Live Feed</span>
        </div>

        {#if loading}
          <div class="py-8 text-center text-xs text-muted font-mono">
            Loading activity stream...
          </div>
        {:else if activities.length === 0}
          <div class="p-6 text-center text-xs text-muted bg-slate-900/30 rounded-xl border border-slate-800 font-mono">
            No recent activity recorded on the swarm.
          </div>
        {:else}
          <div class="space-y-3.5">
            {#each activities as act}
              <div class="flex items-start gap-3 p-3 rounded-xl bg-slate-900/40 border border-slate-800/60 hover:border-slate-700 transition-all">
                <div class="text-base shrink-0 mt-0.5">
                  {getActivityIcon(act.type)}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="text-xs text-main font-sans leading-relaxed">
                    <span class="font-bold font-mono text-emerald-400">@{act.actor}</span>
                    <span class="text-slate-300"> {act.summary}</span>
                  </div>
                  <div class="text-[10px] text-muted font-mono mt-1 flex items-center gap-2">
                    <span>{new Date(act.timestamp).toLocaleTimeString()}</span>
                    <span>•</span>
                    <span class="text-slate-400 truncate">{act.repo_name}</span>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Swarm Radar Telemetry -->
      <SwarmRadar />
    </div>
  </div>
</div>
