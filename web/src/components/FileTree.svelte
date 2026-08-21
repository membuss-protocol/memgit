<script>
  import { Folder, FileText, ChevronRight, FileCode, HardDrive } from 'lucide-svelte';

  let { tree, currentPath, onNavigatePath, onSelectFile } = $props();

  function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  let pathSegments = $derived(currentPath ? currentPath.split('/').filter(Boolean) : []);
</script>

<div class="file-tree card card-elevated">
  <!-- Breadcrumbs Bar -->
  <div class="breadcrumb-bar flex items-center gap-2 text-sm font-mono">
    <button class="breadcrumb-item text-emerald-400 font-semibold" onclick={() => onNavigatePath('')}>
      root
    </button>
    {#each pathSegments as seg, idx}
      <ChevronRight class="w-3.5 h-3.5 text-dim" />
      <button
        class="breadcrumb-item {idx === pathSegments.length - 1 ? 'text-main font-semibold' : 'text-cyan-400'}"
        onclick={() => onNavigatePath(pathSegments.slice(0, idx + 1).join('/'))}
      >
        {seg}
      </button>
    {/each}
  </div>

  <!-- Items Table -->
  <div class="tree-table">
    {#if currentPath}
      <div
        class="tree-row tree-parent-row cursor-pointer"
        onclick={() => {
          const up = pathSegments.slice(0, -1).join('/');
          onNavigatePath(up);
        }}
        role="button"
        tabindex="0"
        onkeydown={(e) => e.key === 'Enter' && onNavigatePath(pathSegments.slice(0, -1).join('/'))}
      >
        <div class="flex items-center gap-2 text-dim font-mono text-xs">
          <Folder class="w-4 h-4 text-cyan-400" />
          <span>..</span>
        </div>
      </div>
    {/if}

    {#if !tree || tree.length === 0}
      <div class="empty-state p-6 text-center text-muted text-sm">
        This directory is empty.
      </div>
    {:else}
      {#each tree as item}
        <div
          class="tree-row cursor-pointer"
          onclick={() => {
            if (item.is_dir) {
              onNavigatePath(item.path);
            } else {
              onSelectFile(item.path);
            }
          }}
          role="button"
          tabindex="0"
          onkeydown={(e) => e.key === 'Enter' && (item.is_dir ? onNavigatePath(item.path) : onSelectFile(item.path))}
        >
          <div class="flex items-center gap-2.5 flex-1">
            {#if item.is_dir}
              <Folder class="w-4 h-4 text-cyan-400 shrink-0" fill="rgba(56, 189, 248, 0.15)" />
            {:else if item.name.endsWith('.md')}
              <FileText class="w-4 h-4 text-emerald-400 shrink-0" />
            {:else}
              <FileCode class="w-4 h-4 text-slate-400 shrink-0" />
            {/if}
            <span class="font-mono text-sm {item.is_dir ? 'font-medium text-main' : 'text-muted'} hover:text-emerald-400 transition-colors">
              {item.name}
            </span>
          </div>

          <div class="flex items-center gap-4 text-xs font-mono text-dim">
            {#if !item.is_dir}
              <span>{formatBytes(item.size)}</span>
            {/if}
            <span class="badge badge-slate text-xs">{item.is_dir ? 'tree' : 'blob'}</span>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .file-tree {
    padding: 0;
    overflow: hidden;
    border: 1px solid var(--border-subtle);
  }

  .breadcrumb-bar {
    padding: 0.75rem 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .breadcrumb-item:hover {
    text-decoration: underline;
  }

  .tree-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.65rem 1.25rem;
    border-bottom: 1px solid rgba(36, 48, 72, 0.5);
    transition: background 0.15s ease;
  }

  .tree-row:last-child {
    border-bottom: none;
  }

  .tree-row:hover {
    background: rgba(30, 41, 59, 0.7);
  }
</style>
