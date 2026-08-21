<script>
  import { X, Copy, Check, Terminal, Radio, Shield, Globe, Server } from 'lucide-svelte';

  let { repo, onClose } = $props();

  let activeTab = $state('https'); // 'https' | 'gateway' | 'membuss'
  let copied = $state(false);

  let cloneCommand = $derived.by(() => {
    if (activeTab === 'https') {
      return `git clone ${repo.clone_https}`;
    }
    if (activeTab === 'gateway') {
      return `git clone ${repo.clone_gateway || ('https://gateway.membuss.dpdns.org/memns/' + repo.memns_name)}`;
    }
    return `git clone ${repo.clone_membuss}`;
  });

  function handleCopy() {
    navigator.clipboard.writeText(cloneCommand);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }
</script>

<div
  class="modal-overlay"
  onclick={onClose}
  onkeydown={(e) => e.key === 'Escape' && onClose()}
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
    <!-- Header -->
    <div class="flex items-center justify-between p-5 border-b border-slate-800">
      <div class="flex items-center gap-2">
        <Terminal class="w-5 h-5 text-emerald-400" />
        <h3 class="text-base font-bold text-main font-mono">Clone {repo.name}</h3>
      </div>
      <button class="btn btn-ghost btn-sm text-muted hover:text-main" onclick={onClose} aria-label="Close modal">
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- Body -->
    <div class="p-6">
      <!-- Protocol Switcher -->
      <div class="protocol-tabs flex gap-1.5 mb-4 p-1 bg-slate-900 rounded-lg border border-slate-800">
        <button
          class="flex-1 py-1.5 px-2 rounded-md text-xs font-semibold font-mono transition-all {activeTab === 'https' ? 'bg-emerald-600 text-white shadow' : 'text-muted hover:text-main'}"
          onclick={() => activeTab = 'https'}
        >
          Git Smart HTTP
        </button>
        <button
          class="flex-1 py-1.5 px-2 rounded-md text-xs font-semibold font-mono transition-all {activeTab === 'gateway' ? 'bg-cyan-600 text-white shadow' : 'text-muted hover:text-main'}"
          onclick={() => activeTab = 'gateway'}
        >
          Membuss Gateway CDN
        </button>
        <button
          class="flex-1 py-1.5 px-2 rounded-md text-xs font-semibold font-mono transition-all {activeTab === 'membuss' ? 'bg-purple-600 text-white shadow' : 'text-muted hover:text-main'}"
          onclick={() => activeTab = 'membuss'}
        >
          P2P Scheme
        </button>
      </div>

      <!-- Command Snippet Box -->
      <div class="command-box flex items-center justify-between gap-3 p-3 bg-slate-950 border border-slate-800 rounded-lg mb-6">
        <span class="font-mono text-xs text-emerald-400 select-all overflow-x-auto whitespace-nowrap">
          {cloneCommand}
        </span>
        <button class="btn btn-secondary btn-sm shrink-0" onclick={handleCopy}>
          {#if copied}
            <Check class="w-3.5 h-3.5 text-emerald-400" />
            <span class="text-emerald-400 text-xs">Copied</span>
          {:else}
            <Copy class="w-3.5 h-3.5 text-muted" />
            <span class="text-xs">Copy</span>
          {/if}
        </button>
      </div>

      <!-- Protocol Explanation -->
      {#if activeTab === 'https'}
        <div class="p-3 bg-slate-900/60 rounded-lg border border-slate-800 text-xs text-muted flex items-start gap-2.5">
          <Globe class="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
          <p>
            Standard Git Smart HTTP protocol. Works out of the box with any official <code>git</code> client. On push, MemGit automatically chunks objects into Merkle DAGs and signs the root MID to MemNS.
          </p>
        </div>
      {:else if activeTab === 'gateway'}
        <div class="p-3 bg-slate-900/60 rounded-lg border border-slate-800 text-xs text-muted flex items-start gap-2.5">
          <Server class="w-4 h-4 text-cyan-400 shrink-0 mt-0.5" />
          <p>
            Direct clone via the <strong>Public Membuss Gateway & CDN</strong> over Dumb HTTP. Any machine in the world can clone your repository straight from any Membuss Gateway without needing the MemGit server running!
          </p>
        </div>
      {:else}
        <div class="p-3 bg-slate-900/60 rounded-lg border border-slate-800 text-xs text-muted flex items-start gap-2.5">
          <Radio class="w-4 h-4 text-purple-400 shrink-0 mt-0.5" />
          <p>
            Native P2P transport using <code>membuss://</code>. Streams chunked Merkle DAG objects directly from swarm peers and Anchor nodes with 10+4 Reed-Solomon erasure coding.
          </p>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-4 bg-slate-900/50 border-t border-slate-800 flex justify-end">
      <button class="btn btn-secondary text-xs" onclick={onClose}>
        Done
      </button>
    </div>
  </div>
</div>

<style>
  .command-box {
    box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.5);
  }
</style>
