<script>
  import { Radio, ShieldCheck, Database, RefreshCw, Users, Server, Zap } from 'lucide-svelte';

  let { swarm, onRefresh } = $props();
  let refreshing = $state(false);

  async function handleRefresh() {
    refreshing = true;
    if (onRefresh) {
      await onRefresh();
    }
    setTimeout(() => {
      refreshing = false;
    }, 600);
  }
</script>

<div class="swarm-radar card card-elevated">
  <!-- Header -->
  <div class="radar-header flex items-center justify-between">
    <div class="flex items-center gap-2">
      <Radio class="w-4 h-4 text-cyan-400" />
      <span class="font-mono text-sm font-semibold text-main">P2P Swarm & Replication Radar</span>
    </div>
    <button class="btn btn-secondary btn-sm" onclick={handleRefresh} disabled={refreshing}>
      <RefreshCw class="w-3.5 h-3.5 text-muted {refreshing ? 'animate-spin' : ''}" />
      <span class="text-xs">Refresh</span>
    </button>
  </div>

  <!-- Metric Grid -->
  <div class="p-6">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <!-- Metric 1: Connected Peers -->
      <div class="metric-card card">
        <div class="flex items-center justify-between text-dim mb-2">
          <span class="text-xs font-semibold">Swarm Peers</span>
          <Users class="w-4 h-4 text-emerald-400" />
        </div>
        <div class="text-2xl font-bold font-mono text-main">{swarm?.connected_peers || 0}</div>
        <div class="text-xs text-muted mt-1">Direct libp2p streams</div>
      </div>

      <!-- Metric 2: DHT Providers -->
      <div class="metric-card card">
        <div class="flex items-center justify-between text-dim mb-2">
          <span class="text-xs font-semibold">DHT Providers</span>
          <Server class="w-4 h-4 text-cyan-400" />
        </div>
        <div class="text-2xl font-bold font-mono text-main">{swarm?.dht_providers || 0}</div>
        <div class="text-xs text-muted mt-1">Mem-DHT announce records</div>
      </div>

      <!-- Metric 3: Erasure Coding -->
      <div class="metric-card card">
        <div class="flex items-center justify-between text-dim mb-2">
          <span class="text-xs font-semibold">Erasure Parity</span>
          <ShieldCheck class="w-4 h-4 text-purple-400" />
        </div>
        <div class="text-2xl font-bold font-mono text-purple-300">10 + 4</div>
        <div class="text-xs text-muted mt-1">Survives 4 node failures</div>
      </div>

      <!-- Metric 4: Anchor State -->
      <div class="metric-card card">
        <div class="flex items-center justify-between text-dim mb-2">
          <span class="text-xs font-semibold">Anchor Backup</span>
          <Zap class="w-4 h-4 text-amber-400" />
        </div>
        <div class="text-2xl font-bold font-mono text-amber-300">
          {swarm?.anchor_synced ? 'Active' : 'Standby'}
        </div>
        <div class="text-xs text-muted mt-1">Permanent archival layer</div>
      </div>
    </div>

    <!-- Active Swarm Peers Table -->
    <div class="peers-box card">
      <div class="text-xs font-semibold text-main mb-3 font-mono">Connected Swarm Nodes:</div>
      {#if !swarm?.peers || swarm.peers.length === 0}
        <div class="p-4 text-center text-xs text-dim">
          No external peers currently streaming this repository shard. Local BadgerDB cache active.
        </div>
      {:else}
        <div class="space-y-2">
          {#each swarm.peers as peer}
            <div class="peer-row flex items-center justify-between p-2.5 bg-slate-900/60 rounded-lg border border-slate-800 font-mono text-xs">
              <div class="flex items-center gap-2">
                <span class="status-dot"></span>
                <span class="text-cyan-300">{peer.id}</span>
              </div>
              <span class="badge badge-emerald text-xs">{peer.latency || '12ms'}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .swarm-radar {
    padding: 0;
    overflow: hidden;
  }

  .radar-header {
    padding: 0.75rem 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .metric-card {
    background: #0d121f;
    border: 1px solid var(--border-subtle);
    padding: 1rem;
  }

  .peers-box {
    background: #090d16;
    border: 1px solid var(--border-subtle);
    padding: 1rem;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #22c55e;
    box-shadow: 0 0 6px #22c55e;
  }
</style>
