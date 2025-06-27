<script>
  import logo from "$lib/assets/images/logo-universal.png";
  import { Greet } from "$lib/wailsjs/go/main/App.js";

  let resultText = $state("Please enter your name below 👇");
  let name = $state("");

  function greet() {
    Greet(name).then((result) => (resultText = result));
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-2xl mx-auto text-center space-y-6">
    <h1 class="text-4xl font-bold text-primary">Welcome to the Unofficial Wails.io SvelteKit Template!</h1>
    <p class="text-muted-foreground">
      Visit <a href="https://kit.svelte.dev" class="text-primary hover:underline">kit.svelte.dev</a>
      to read the documentation
    </p>
    <img alt="Wails logo" src={logo} class="w-64 h-64 mx-auto object-contain" />
    <div class="text-lg font-medium text-card-foreground bg-card p-4 rounded-lg border">{resultText}</div>
    <div class="flex gap-4 justify-center items-center">
      <input
        autocomplete="off"
        bind:value={name}
        class="px-4 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        id="name"
        type="text"
        placeholder="Enter your name"
      />
      <button 
        class="px-6 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
        onclick={greet}
      >
        Greet
      </button>
    </div>
  </div>
</main>
