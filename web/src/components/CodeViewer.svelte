<script>
  import { onMount } from 'svelte';
  import { Copy, Check, Download, FileCode, ArrowLeft } from 'lucide-svelte';
  import Prism from 'prismjs';
  import 'prismjs/components/prism-go';
  import 'prismjs/components/prism-javascript';
  import 'prismjs/components/prism-typescript';
  import 'prismjs/components/prism-json';
  import 'prismjs/components/prism-markdown';
  import 'prismjs/components/prism-bash';
  import 'prismjs/components/prism-yaml';
  import 'prismjs/components/prism-rust';
  import 'prismjs/components/prism-python';

  let { blob, filePath, onBack } = $props();
  let copied = $state(false);

  function getLanguage(path) {
    if (!path) return 'javascript';
    const ext = path.split('.').pop().toLowerCase();
    const map = {
      go: 'go',
      js: 'javascript',
      jsx: 'javascript',
      ts: 'typescript',
      tsx: 'typescript',
      json: 'json',
      md: 'markdown',
      sh: 'bash',
      bash: 'bash',
      yaml: 'yaml',
      yml: 'yaml',
      rs: 'rust',
      py: 'python',
    };
    return map[ext] || 'javascript';
  }

  function handleCopy() {
    if (!blob?.content) return;
    navigator.clipboard.writeText(blob.content);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  function highlightCode(content, path) {
    if (!content) return '';
    const lang = getLanguage(path);
    if (Prism.languages[lang]) {
      return Prism.highlight(content, Prism.languages[lang], lang);
    }
    return content.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  let lines = $derived(blob?.content ? blob.content.split('\n') : []);
</script>

<div class="code-viewer card card-elevated">
  <!-- Header Bar -->
  <div class="viewer-header flex items-center justify-between">
    <div class="flex items-center gap-3">
      <button class="btn btn-ghost btn-sm" onclick={onBack} title="Back to Directory">
        <ArrowLeft class="w-4 h-4 text-emerald-400" />
      </button>
      <div class="flex items-center gap-2">
        <FileCode class="w-4 h-4 text-cyan-400" />
        <span class="font-mono text-sm font-semibold text-main">{filePath}</span>
        <span class="badge badge-slate text-xs font-mono">{lines.length} lines</span>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2">
      <button class="btn btn-secondary btn-sm" onclick={handleCopy}>
        {#if copied}
          <Check class="w-3.5 h-3.5 text-emerald-400" />
          <span class="text-emerald-400">Copied</span>
        {:else}
          <Copy class="w-3.5 h-3.5" />
          <span>Copy</span>
        {/if}
      </button>
    </div>
  </div>

  <!-- Code Body with Line Numbers -->
  <div class="code-container font-mono text-xs">
    <div class="line-numbers-col">
      {#each lines as _, idx}
        <span class="line-num">{idx + 1}</span>
      {/each}
    </div>

    <div class="code-lines-col">
      <pre class="language-{getLanguage(filePath)}"><code>{@html highlightCode(blob?.content || '', filePath)}</code></pre>
    </div>
  </div>
</div>

<style>
  .code-viewer {
    padding: 0;
    overflow: hidden;
  }

  .viewer-header {
    padding: 0.75rem 1.25rem;
    background: #0d121f;
    border-bottom: 1px solid var(--border-subtle);
  }

  .code-container {
    display: flex;
    background: #080c14;
    overflow-x: auto;
  }

  .line-numbers-col {
    display: flex;
    flex-direction: column;
    padding: 1rem 0.75rem;
    background: #090d17;
    border-right: 1px solid var(--border-subtle);
    user-select: none;
    text-align: right;
    color: var(--text-dim);
  }

  .line-num {
    height: 1.6em;
    line-height: 1.6em;
    min-width: 2.2rem;
  }

  .code-lines-col {
    flex: 1;
    overflow-x: auto;
  }

  .code-lines-col pre {
    margin: 0;
    padding: 1rem !important;
    background: transparent !important;
    border: none !important;
    line-height: 1.6em !important;
  }
</style>
