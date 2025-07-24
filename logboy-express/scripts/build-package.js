import {
  copyFileSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from "fs";

mkdirSync("build/cjs", { recursive: true });
mkdirSync("build/esm", { recursive: true });
mkdirSync("build/protobuf", { recursive: true });

copyFileSync("protobuf/log.proto", "build/protobuf/log.proto");
copyFileSync("protobuf/metrics.proto", "build/protobuf/metrics.proto");

writeFileSync(
  "build/cjs/package.json",
  JSON.stringify(
    {
      type: "commonjs",
    },
    null,
    2
  )
);

writeFileSync(
  "build/esm/package.json",
  JSON.stringify(
    {
      type: "module",
    },
    null,
    2
  )
);

const pkg = JSON.parse(readFileSync("package.json", "utf8"));

const buildPkg = {
  name: pkg.name,
  version: pkg.version,
  description: pkg.description,
  author: pkg.author,
  license: pkg.license,
  keywords: pkg.keywords,
  repository: pkg.repository,
  bugs: pkg.bugs,
  homepage: pkg.homepage,
  main: "./cjs/index.js",
  module: "./esm/index.js",
  types: "./types/index.d.ts",
  exports: {
    ".": {
      import: "./esm/index.js",
      require: "./cjs/index.js",
      types: "./types/index.d.ts",
    },
  },
  files: ["cjs", "esm", "types", "protobuf"],
  dependencies: pkg.dependencies,
  peerDependencies: pkg.peerDependencies,
  engines: pkg.engines,
};

writeFileSync("build/package.json", JSON.stringify(buildPkg, null, 2));
