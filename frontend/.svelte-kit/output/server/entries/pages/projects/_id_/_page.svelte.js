import { A as getContext, $ as store_get, P as copy_payload, Q as assign_payload, a0 as unsubscribe_stores, y as pop, w as push } from "../../../../chunks/index.js";
import { B as Button } from "../../../../chunks/button.js";
import "clsx";
import { g as goto } from "../../../../chunks/client.js";
const getStores = () => {
  const stores$1 = getContext("__svelte__");
  return {
    /** @type {typeof page} */
    page: {
      subscribe: stores$1.page.subscribe
    },
    /** @type {typeof navigating} */
    navigating: {
      subscribe: stores$1.navigating.subscribe
    },
    /** @type {typeof updated} */
    updated: stores$1.updated
  };
};
const page = {
  subscribe(fn) {
    const store = getStores().page;
    return store.subscribe(fn);
  }
};
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  parseInt(store_get($$store_subs ??= {}, "$page", page).params.id);
  function goBack() {
    goto();
  }
  let $$settled = true;
  let $$inner_payload;
  function $$render_inner($$payload2) {
    $$payload2.out += `<main class="min-h-screen bg-background text-foreground p-8"><div class="max-w-4xl mx-auto space-y-6"><div class="flex items-center gap-4">`;
    Button($$payload2, {
      variant: "outline",
      onclick: goBack,
      class: "flex items-center gap-2",
      children: ($$payload3) => {
        $$payload3.out += `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg> Back to Projects`;
      },
      $$slots: { default: true }
    });
    $$payload2.out += `<!----></div> `;
    {
      $$payload2.out += "<!--[!-->";
    }
    $$payload2.out += `<!--]--> `;
    {
      $$payload2.out += "<!--[2-->";
      $$payload2.out += `<div class="text-center py-12 text-muted-foreground"><p class="text-lg">Project not found</p> <p class="text-sm">The project you're looking for doesn't exist</p> `;
      Button($$payload2, {
        class: "mt-4",
        onclick: goBack,
        children: ($$payload3) => {
          $$payload3.out += `<!---->Go Back`;
        },
        $$slots: { default: true }
      });
      $$payload2.out += `<!----></div>`;
    }
    $$payload2.out += `<!--]--></div></main>`;
  }
  do {
    $$settled = true;
    $$inner_payload = copy_payload($$payload);
    $$render_inner($$inner_payload);
  } while (!$$settled);
  assign_payload($$payload, $$inner_payload);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
