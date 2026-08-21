<script>
  import { marked } from 'marked';
  import { BookOpen } from 'lucide-svelte';

  let { content, title = 'README.md' } = $props();

  let html = $derived(content ? marked.parse(content) : '');
</script>

<div class="card card-elevated markdown-card">
  <div class="markdown-header flex items-center gap-2">
    <BookOpen class="w-4 h-4 text-emerald-400" />
    <span class="font-mono text-sm font-semibold text-main">{title}</span>
  </div>

  <div class="markdown-body p-6">
    {@html html}
  </div>
</div>

<style>
  .markdown-card {
    padding: 0;
    overflow: hidden;
  }

  .markdown-header {
    padding: 0.75rem 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .markdown-body {
    padding: 1.5rem;
  }
</style>
