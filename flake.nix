{
  description = "Free Mind: block distracting websites and stay focused";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        free-mind = pkgs.callPackage ./packaging/nix/package.nix { };
        default = free-mind;
      });

      # NixOS module: services.free-mind.{enable,user}.
      nixosModules.default = import ./packaging/nix/module.nix self;
    };
}
