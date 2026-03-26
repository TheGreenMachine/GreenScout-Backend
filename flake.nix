{
  description = "GreenMachine Backend";

  inputs = {
    nixpkgs.url = "github:Nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShells.default = pkgs.mkShell {
          hardeningDisable = [ "fortify" ];
          buildInputs = with pkgs;
            [
              go
              gopls
              gotools
              go-tools
            ]
            ++ systemDeps;

          shellHook = ''
            export CGO_ENABLED=1
          '';
        };
      }
    );
}
