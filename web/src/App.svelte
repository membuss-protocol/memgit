<script>
  import { onMount } from 'svelte';
  import { api } from './api/client.js';
  import Navbar from './components/Navbar.svelte';
  import Dashboard from './pages/Dashboard.svelte';
  import ExplorePage from './pages/ExplorePage.svelte';
  import UserProfilePage from './pages/UserProfilePage.svelte';
  import RepoView from './pages/RepoView.svelte';
  import NewRepoModal from './components/NewRepoModal.svelte';
  import ProfileModal from './components/ProfileModal.svelte';

  let currentRoute = $state('dashboard'); // 'dashboard' | 'explore' | 'repo' | 'user'
  let routeParams = $state({});
  let currentUser = $state(null);
  let showNewRepoModal = $state(false);
  let showProfileModal = $state(false);

  function parseHash() {
    const hash = window.location.hash.slice(1);
    if (!hash || hash === '/' || hash === 'dashboard') {
      currentRoute = 'dashboard';
      routeParams = {};
      return;
    }

    if (hash.startsWith('explore')) {
      currentRoute = 'explore';
      const parts = hash.split('?');
      if (parts[1]) {
        const search = new URLSearchParams(parts[1]);
        routeParams = { filter: search.get('filter') || 'all', q: search.get('q') || '' };
      } else {
        routeParams = { filter: 'all', q: '' };
      }
      return;
    }

    if (hash.startsWith('user/')) {
      currentRoute = 'user';
      routeParams = { username: decodeURIComponent(hash.slice(5)) };
      return;
    }

    if (hash.startsWith('repo/')) {
      currentRoute = 'repo';
      routeParams = { repoName: decodeURIComponent(hash.slice(5)) };
      return;
    }

    // Default to repo if legacy hash
    currentRoute = 'repo';
    routeParams = { repoName: decodeURIComponent(hash) };
  }

  onMount(async () => {
    window.addEventListener('hashchange', parseHash);
    parseHash();

    try {
      currentUser = await api.getCurrentUser();
    } catch (e) {
      console.error('Failed to load current user', e);
    }
  });

  function navigate(route, params = {}) {
    currentRoute = route;
    routeParams = params;

    if (route === 'dashboard') {
      window.location.hash = '';
    } else if (route === 'explore') {
      const q = params.q ? `&q=${encodeURIComponent(params.q)}` : '';
      const filter = params.filter ? `filter=${params.filter}` : 'filter=all';
      window.location.hash = `explore?${filter}${q}`;
    } else if (route === 'user') {
      window.location.hash = `user/${encodeURIComponent(params.username)}`;
    } else if (route === 'repo') {
      window.location.hash = `repo/${encodeURIComponent(params.repoName)}`;
    }
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  async function handleCreateRepo(data) {
    try {
      const newRepo = await api.createRepo(data);
      showNewRepoModal = false;
      navigate('repo', { repoName: newRepo.name });
    } catch (err) {
      alert('Failed to create repository: ' + err.message);
    }
  }
</script>

<div class="app-layout min-h-screen flex flex-col bg-base text-main selection:bg-emerald-500/30 selection:text-emerald-200">
  <!-- Top Global Navbar -->
  <Navbar
    {currentRoute}
    {currentUser}
    onNavigate={navigate}
    onNewRepo={() => showNewRepoModal = true}
    onOpenProfile={() => showProfileModal = true}
  />

  <!-- Main Multi-Page Route Container -->
  <main class="flex-1">
    {#if currentRoute === 'dashboard'}
      <Dashboard
        {currentUser}
        onSelectRepo={(name) => navigate('repo', { repoName: name })}
        onNewRepo={() => showNewRepoModal = true}
        onNavigate={navigate}
      />
    {:else if currentRoute === 'explore'}
      <ExplorePage
        initialFilter={routeParams.filter || 'all'}
        initialQuery={routeParams.q || ''}
        onSelectRepo={(name) => navigate('repo', { repoName: name })}
        onNewRepo={() => showNewRepoModal = true}
      />
    {:else if currentRoute === 'user'}
      <UserProfilePage
        username={routeParams.username}
        onSelectRepo={(name) => navigate('repo', { repoName: name })}
        onNewRepo={() => showNewRepoModal = true}
      />
    {:else if currentRoute === 'repo'}
      <RepoView
        repoName={routeParams.repoName}
        onGoBack={() => navigate('explore')}
        onNavigate={navigate}
      />
    {/if}
  </main>

  <!-- Global Footer -->
  <footer class="py-8 border-t border-slate-800/80 bg-surface/40 text-xs text-muted">
    <div class="max-w-7xl mx-auto px-4 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4">
      <div class="flex items-center gap-2 font-mono">
        <span class="font-bold text-main">MemGit</span>
        <span class="text-slate-600">•</span>
        <span>Decentralized Git on Membuss Network</span>
      </div>
      <div class="flex items-center gap-4 font-mono text-[11px] text-muted">
        <span>Libp2p Swarm</span>
        <span>•</span>
        <span>Reed-Solomon 10+4 Parity</span>
        <span>•</span>
        <span>Ed25519 MemNS</span>
      </div>
    </div>
  </footer>
</div>

<!-- Modals -->
{#if showNewRepoModal}
  <NewRepoModal
    onClose={() => showNewRepoModal = false}
    onCreate={handleCreateRepo}
  />
{/if}

{#if showProfileModal}
  <ProfileModal
    {currentUser}
    onClose={() => showProfileModal = false}
    onProfileUpdated={(updated) => currentUser = updated}
  />
{/if}
