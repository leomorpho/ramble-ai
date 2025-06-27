import { e as escape_html } from "../../chunks/escaping.js";
import { c as pop, p as push } from "../../chunks/index.js";
const replacements = {
  translate: /* @__PURE__ */ new Map([
    [true, "yes"],
    [false, "no"]
  ])
};
function attr(name, value, is_boolean = false) {
  if (value == null || !value && is_boolean) return "";
  const normalized = name in replacements && replacements[name].get(value) || value;
  const assignment = is_boolean ? "" : `="${escape_html(normalized, true)}"`;
  return ` ${name}${assignment}`;
}
const logo = "/_app/immutable/assets/logo-universal.Dm-wv4TN.png";
function _page($$payload, $$props) {
  push();
  let resultText = "Please enter your name below 👇";
  let name = "";
  $$payload.out += `<main><h1>Welcome to the Unofficial Wails.io SvelteKit Template!</h1> <p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p> <img alt="Wails logo" id="logo"${attr("src", logo)} class="svelte-140lxdh"/> <div class="result svelte-140lxdh" id="result">${escape_html(resultText)}</div> <div class="input-box svelte-140lxdh" id="input"><input autocomplete="off"${attr("value", name)} class="input svelte-140lxdh" id="name" type="text"/> <button class="btn svelte-140lxdh">Greet</button></div></main>`;
  pop();
}
export {
  _page as default
};
