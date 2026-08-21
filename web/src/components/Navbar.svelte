<script>
  import { GitBranch, Plus, Search, Terminal, Radio, Shield, User, Globe, Compass, Activity, ChevronDown } from 'lucide-svelte';

  let { currentRoute, currentUser, onNavigate, onNewRepo, onOpenProfile } = $props();

  let searchQuery = $state('');
  let showUserMenu = $state(false);

  function handleSearch(e) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      onNavigate('explore', { q: searchQuery.trim() });
    }
  }

  function getAvatarColor(name) {
    const colors = ['bg-emerald-600', 'bg-cyan-600', 'bg-purple-600', 'bg-blue-600', 'bg-amber-600'];
    let hash = 0;
    for (let i = 0; i < (name || '').length; i++) hash += name.charCodeAt(i);
    return colors[Math.abs(hash) % colors.length];
  }
</script>

<header class="sticky top-0 z-40 bg-surface/90 backdrop-blur border-b border-slate-800/80 px-4 lg:px-8 py-3">
  <div class="max-w-7xl mx-auto flex items-center justify-between gap-4">
    <!-- Brand & Logo -->
    <div class="flex items-center gap-6">
      <button class="flex items-center gap-3 text-left group" onclick={() => onNavigate('dashboard')}>
        <div class="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center group-hover:border-emerald-400/60 transition-colors">
          <GitBranch class="w-4 h-4 text-emerald-400" />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="font-bold text-sm tracking-tight text-main font-mono">MemGit</span>
            <span class="badge badge-cyan text-[10px] py-0 px-1.5 font-mono">P2P</span>
          </div>
          <span class="text-[10px] text-muted block -mt-0.5">Decentralized Git on Membuss</span>
        </div>
      </button>

      <!-- Navigation Tabs -->
      <nav class="hidden md:flex items-center gap-1">
        <button
          class="nav-tab {currentRoute === 'dashboard' ? 'nav-tab-active' : ''}"
          onclick={() => onNavigate('dashboard')}
        >
          Dashboard
        </button>
        <button
          class="nav-tab {currentRoute === 'explore' ? 'nav-tab-active' : ''}"
          onclick={() => onNavigate('explore')}
        >
          <Compass class="w-3.5 h-3.5 inline mr-1 text-cyan-400" />
          Explore Swarm
        </button>
      </nav>
    </div>

    <!-- Search & User Actions -->
    <div class="flex items-center gap-3">
      <!-- Search Input -->
      <div class="relative hidden sm:block w-56 lg:w-72">
        <Search class="w-3.5 h-3.5 text-muted absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          bind:value={searchQuery}
          onkeydown={handleSearch}
          placeholder="Search repositories, topics, users..."
          class="input-field pl-9 pr-3 py-1.5 text-xs w-full font-sans placeholder:text-muted/70 bg-slate-950/60"
        />
      </div>

      <!-- New Repository CTA -->
      <button class="btn btn-primary btn-sm flex items-center gap-1.5" onclick={onNewRepo}>
        <Plus class="w-3.5 h-3.5" />
        <span class="hidden sm:inline">New Repository</span>
      </button>

      <!-- User Profile Menu -->
      <div class="relative">
        <button
          class="flex items-center gap-2 p-1 pl-2 pr-2.5 rounded-lg border border-slate-800 bg-slate-900/60 hover:border-slate-700 transition-colors"
          onclick={() => showUserMenu = !showUserMenu}
        >
          <div class="w-6 h-6 rounded-full {getAvatarColor(currentUser?.username)} text-white text-[11px] font-bold flex items-center justify-center font-mono uppercase">
            {(currentUser?.username || 'D')[0]}
          </div>
          <span class="text-xs font-mono text-main hidden md:inline">
            @{currentUser?.username || 'developer'}
          </span>
          <ChevronDown class="w-3 h-3 text-muted" />
        </button>

        {#if showUserMenu}
          <div
            class="user-dropdown absolute right-0 mt-2 w-56 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl py-1.5 z-50 animate-in fade-in zoom-in-95 duration-100"
            onclick={() => showUserMenu = false}
            onkeydown={(e) => e.key === 'Escape' && (showUserMenu = false)}
            role="menu"
            tabindex="0"
          >
            <div class="px-3.5 py-2.5 border-b border-slate-800/80">
              <div class="text-xs font-bold text-main">{currentUser?.display_name || 'Membuss Contributor'}</div>
              <div class="text-[11px] text-muted font-mono truncate">@{currentUser?.username || 'developer'}</div>
              <div class="mt-1 flex items-center gap-1 text-[10px] text-emerald-400 font-mono">
                <Shield class="w-3 h-3" /> Ed25519 Signed
              </div>
            </div>

            <button
              class="dropdown-item w-full text-left"
              onclick={() => onNavigate('user', { username: currentUser?.username })}
            >
              <User class="w-3.5 h-3.5 text-muted" />
              <span>Your Profile</span>
            </button>

            <button
              class="dropdown-item w-full text-left"
              onclick={onOpenProfile}
            >
              <Terminal class="w-3.5 h-3.5 text-muted" />
              <span>Identity & Keys</span>
            </button>

            <button
              class="dropdown-item w-full text-left"
              onclick={() => onNavigate('explore', { filter: 'starred' })}
            >
              <Globe class="w-3.5 h-3.5 text-muted" />
              <span>Your Starred Repos</span>
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
</header>

<style>
  .nav-tab {
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-text-muted);
    border-radius: 0.375rem;
    transition: all 180ms ease;
  }
  .nav-tab:hover {
    color: var(--color-text-main);
    background-color: rgba(30, 41, 59, 0.5);
  }
  .nav-tab-active {
    color: var(--color-text-main);
    background-color: rgba(30, 41, 59, 0.8);
    font-weight: 600;
  }
  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5rem 0.875rem;
    font-size: 0.75rem;
    color: var(--color-text-muted);
    transition: all 150ms ease;
  }
  .dropdown-item:hover {
    color: var(--color-text-main);
    background-color: rgba(51, 65, 85, 0.4);
  }
</style>
