{
  description = "Manga Helper tool";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";

    devtools.url     = "github:nanoteck137/devtools";
    devtools.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, devtools, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        version = pkgs.lib.strings.fileContents "${self}/version";
        fullVersion = ''${version}-${self.dirtyShortRev or self.shortRev or "dirty"}'';

        app = pkgs.buildGoModule {
          pname = "patrat";
          version = fullVersion;
          src = ./.;

          ldflags = [
            "-X github.com/nanoteck137/patrat/cmd.Version=${version}"
            "-X github.com/nanoteck137/patrat/cmd.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          ];

          vendorHash = "sha256-oqL3RObx+i3k1x09uQW1bKyhGQYGwXaFlGZx+Kukeq0=";
        };

        tools = devtools.packages.${system};
      in
      {
        packages.default = app;
        packages.patrat = app;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            tools.publishVersion
          ];
        };
      }
    );
}
