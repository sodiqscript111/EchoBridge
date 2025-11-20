<script>
  import "../app.css";
  import Navbar from "$lib/components/Navbar.svelte";
</script>

<div class="min-h-screen bg-black text-white selection:bg-apple-blue selection:text-white">
  <Navbar />
  <slot />
</div>
