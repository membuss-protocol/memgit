<script>
  import { onMount } from 'svelte';
  import { User, Mail, Calendar, Shield, Star, GitFork, BookOpen, Terminal } from 'lucide-svelte';
  import { api } from '../api/client.js';
  import RepoCard from '../components/RepoCard.svelte';

  let { username, onSelectRepo, onNewRepo } = $props();

  let profileData = $state(null);
  let loading = $state(true);
  let error = $state('');
  let activeTab = $state('repos'); // 'repos' | 'starred'

  async function loadProfile() {
    loading = true;
    error = '';
    try {
      const data = await api.getUserProfile(username);
      profileData = data;
    } catch (err) {
      error = err.message || 'Failed to load user profile';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadProfile();
  });

  function getAvatarColor(name) {
    const colors = ['bg-emerald-600', 'bg-cyan-600', 'bg-purple-600', 'bg-blue-600', 'bg-amber-600'];
    let hash = 0;
    for (let i = 0; i < (name || '').length; i++) hash += name.charCodeAt(i);
    return colors[Math.abs(hash) % colors.length];
  }
</script>

<div class="max-w-7xl mx-auto px-4 lg:px-8 py-8">
  {#if loading}
    <div class="p-12 text-center text-muted font-mono text-xs">
      <div class="animate-spin w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full mx-auto mb-3"></div>
      Loading developer profile from Membuss DHT...
    </div>
  {:else if error}
    <div class="p-6 bg-red-950/30 border border-red-800/40 rounded-xl text-xs text-red-400 text-center">
      {error}
    </div>
  {:else if profileData}
    <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
      <!-- Left Column: User Profile Card -->
      <div class="lg:col-span-1 space-y-5">
        <div class="card p-6 text-center">
          <div class="w-24 h-24 rounded-2xl {getAvatarColor(profileData.user.username)} text-white text-3xl font-extrabold flex items-center justify-center font-mono mx-auto mb-4 shadow-xl uppercase">
            {profileData.user.username[0]}
          </div>

          <h2 class="text-lg font-bold text-main">{profileData.user.display_name || profileData.user.username}</h2>
          <p class="text-xs font-mono text-muted mb-4">@{profileData.user.username}</p>

          {#if profileData.user.bio}
            <p class="text-xs text-slate-300 leading-relaxed text-left mb-5 p-3 bg-slate-900/60 rounded-lg border border-slate-800">
              {profileData.user.bio}
            </p>
          {/if}

          <div class="space-y-2 text-xs text-muted text-left font-mono pt-4 border-t border-slate-800/80">
            {#if profileData.user.email}
              <div class="flex items-center gap-2 truncate">
                <Mail class="w-3.5 h-3.5 text-muted shrink-0" />
                <span class="truncate">{profileData.user.email}</span>
              </div>
            {/if}
            <div class="flex items-center gap-2">
              <Calendar class="w-3.5 h-3.5 text-muted shrink-0" />
              <span>Joined {new Date(profileData.user.joined_at).toLocaleDateString()}</span>
            </div>
            <div class="flex items-center gap-2 text-emerald-400">
              <Shield class="w-3.5 h-3.5 shrink-0" />
              <span>Ed25519 Verified</span>
            </div>
          </div>
        </div>

        <!-- Cryptographic Public Key Card -->
        <div class="card p-4">
          <div class="flex items-center gap-2 text-xs font-bold text-cyan-400 font-mono mb-2">
            <Shield class="w-3.5 h-3.5" /> Public Key Identity
          </div>
          <p class="text-[11px] font-mono text-muted break-all select-all p-2 bg-slate-950 rounded border border-slate-800">
            {profileData.user.public_key}
          </p>
        </div>
      </div>

      <!-- Right Column: Tabs (Repositories & Starred) -->
      <div class="lg:col-span-3 space-y-6">
        <!-- Sub-Navigation -->
        <div class="flex items-center gap-2 border-b border-slate-800 pb-3">
          <button
            class="px-4 py-1.5 rounded-lg text-xs font-semibold font-mono transition-all {activeTab === 'repos' ? 'bg-emerald-600/20 text-emerald-400 border border-emerald-500/30' : 'text-muted hover:text-main'}"
            onclick={() => activeTab = 'repos'}
          >
            <BookOpen class="w-3.5 h-3.5 inline mr-1" />
            Repositories ({profileData.repositories.length})
          </button>
          <button
            class="px-4 py-1.5 rounded-lg text-xs font-semibold font-mono transition-all {activeTab === 'starred' ? 'bg-amber-600/20 text-amber-400 border border-amber-500/30' : 'text-muted hover:text-main'}"
            onclick={() => activeTab = 'starred'}
          >
            <Star class="w-3.5 h-3.5 inline mr-1" />
            Starred ({profileData.starred_repos.length})
          </button>
        </div>

        <!-- Tab Content -->
        {#if activeTab === 'repos'}
          {#if profileData.repositories.length === 0}
            <div class="p-12 text-center bg-slate-900/30 border border-slate-800 rounded-xl text-muted text-xs font-mono">
              @{profileData.user.username} hasn't published any public repositories yet.
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              {#each profileData.repositories as repo}
                <div class="card p-4 flex flex-col justify-between hover:border-slate-700 transition-all">
                  <div>
                    <button
                      class="text-left font-mono font-bold text-sm text-main hover:text-emerald-400 transition-colors mb-1.5 truncate block w-full"
                      onclick={() => onSelectRepo(repo.name)}
                    >
                      {repo.name}
                    </button>
                    <p class="text-xs text-muted leading-relaxed line-clamp-2 mb-3">
                      {repo.description || 'No description provided.'}
                    </p>
                  </div>
                  <div class="flex items-center justify-between text-xs text-muted font-mono pt-3 border-t border-slate-800/80">
                    <span class="text-[11px]">Updated {new Date(repo.updated_at).toLocaleDateString()}</span>
                    <div class="flex items-center gap-3">
                      <span class="flex items-center gap-1"><Star class="w-3 h-3 text-amber-400" /> {repo.star_count || 0}</span>
                      <span class="flex items-center gap-1"><GitFork class="w-3 h-3 text-cyan-400" /> {repo.fork_count || 0}</span>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {:else}
          {#if profileData.starred_repos.length === 0}
            <div class="p-12 text-center bg-slate-900/30 border border-slate-800 rounded-xl text-muted text-xs font-mono">
              No starred repositories yet.
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              {#each profileData.starred_repos as repo}
                <div class="card p-4 flex flex-col justify-between hover:border-slate-700 transition-all">
                  <div>
                    <button
                      class="text-left font-mono font-bold text-sm text-main hover:text-emerald-400 transition-colors mb-1.5 truncate block w-full"
                      onclick={() => onSelectRepo(repo.name)}
                    >
                      @{repo.owner || 'dev'}/{repo.name}
                    </button>
                    <p class="text-xs text-muted leading-relaxed line-clamp-2 mb-3">
                      {repo.description || 'No description provided.'}
                    </p>
                  </div>
                  <div class="flex items-center justify-between text-xs text-muted font-mono pt-3 border-t border-slate-800/80">
                    <span class="text-[11px]">Updated {new Date(repo.updated_at).toLocaleDateString()}</span>
                    <div class="flex items-center gap-3">
                      <span class="flex items-center gap-1"><Star class="w-3 h-3 text-amber-400 fill-amber-400" /> {repo.star_count || 0}</span>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</div>
