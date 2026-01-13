{
  inputs = {
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };
  };

  outputs =
    {
      self,
      flake-utils,
      nixpkgs,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        mcp-picker = pkgs.buildGoModule {
          pname = "mcp-picker";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-tl9fuQ/84mwarijoNJfPt4uF9C3oPIDvMYmVDSzge8A=";
        };
      in
      {
        packages = {
          default = mcp-picker;
          inherit mcp-picker;
        };

        devShells.default = pkgs.mkShell ({
          buildInputs = with pkgs; [
            go_1_25
            gopls
            nixfmt
          ];
        });
      }
    );
}
