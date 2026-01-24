import { c as create_ssr_component, v as validate_component } from "./ssr.js";
import { I as Icon } from "./Icon.js";
const Log_in = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"
      }
    ],
    ["polyline", { "points": "10 17 15 12 10 7" }],
    [
      "line",
      {
        "x1": "15",
        "x2": "3",
        "y1": "12",
        "y2": "12"
      }
    ]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "log-in" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
export {
  Log_in as L
};
