<script>
  import { Star, GitFork, BookOpen, Key, Shield, Radio, ArrowUpRight } from 'lucide-svelte';

  let { repo, onSelect, onStar } = $props();

  function handleStar(e) {
    e.stopPropagation();
    onStar(repo.name);
  }
</script>

<div class="card card-interactive repo-card" onclick={() => onSelect(repo.name)} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && onSelect(repo.name)}>
  <div class="flex items-start justify-between gap-3 mb-2">
    <div class="flex items-center gap-2">
      <div class="repo-icon">
        <BookOpen class="w-4 h-4 text-emerald-400" />
      </div>
      <h3 class="text-base font-semibold text-main hover:text-emerald-400 transition-colors font-mono">
        {repo.name}
      </h3>
      <span class="badge {repo.is_private ? 'badge-amber' : 'badge-slate'}">
        {repo.is_private ? 'Private' : 'Public'}
      </span>
    </div>

    <!-- Star button -->
    <button class="btn btn-secondary btn-sm star-btn" onclick={handleStar} title="Star Repository">
      <Star class="w-3.5 h-3.5 text-amber-400" fill={repo.star_count > 0 ? '#fbbf24' : 'none'} />
      <span class="font-mono text-xs">{repo.star_count || 0}</span>
    </button>
  </div>

  <p class="text-sm text-muted mb-4 line-clamp-2">
    {repo.description || 'Decentralized repository with cryptographic verification and Reed-Solomon erasure coding.'}
  </p>

  <!-- MemNS & Cryptographic Identity -->
  <div class="crypto-meta flex flex-wrap gap-2 mb-4">
    <div class="meta-pill" title="Cryptographic MemNS Ownership Pointer">
      <Key class="w-3 h-3 text-cyan-400" />
      <span class="font-mono text-xs text-cyan-300">/memns/{repo.memns_name.slice(0, 16)}...</span>
    </div>

    {#if repo.latest_mid}
      <div class="meta-pill" title="Current Merkle DAG Root MID">
        <Shield class="w-3 h-3 text-emerald-400" />
        <span class="font-mono text-xs text-emerald-300">{repo.latest_mid.slice(0, 14)}...</span>
      </div>
    {/if}
  </div>

  <!-- Bottom Row -->
  <div class="flex items-center justify-between text-xs text-dim pt-3 border-t border-slate-800">
    <div class="flex items-center gap-4">
      <span class="flex items-center gap-1">
        <GitFork class="w-3.5 h-3.5" />
        <span>{repo.fork_count || 0}</span>
      </span>
      <span class="badge badge-emerald text-xs">10+4 Parity</span>
    </div>
    <span>Updated {new Date(repo.updated_at).toLocaleDateString()}</span>
  </div>
</div>

<style>
  .repo-card {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    height: 100%;
  }

  .repo-icon {
    width: 28px;
    height: 28px;
    border-radius: 6px;
    background: rgba(34, 197, 94, 0.1);
    border: 1px solid rgba(34, 197, 94, 0.25);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .crypto-meta {
    margin-top: auto;
  }

  .meta-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.25rem 0.6rem;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: 6px;
  }

  .star-btn:hover {
    border-color: #fbbf24;
  }
</style>
