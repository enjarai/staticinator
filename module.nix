{
  packages
}: (
  {
    lib,
    pkgs,
    config,
    ...
  }:

  let
    inherit (lib)
      mkEnableOption
      mkIf
      mkOption
      optionalAttrs
      optional
      mkPackageOption;
    inherit (lib.types)
      bool
      path
      str
      submodule
      number
      array
      listOf;

    dataPath = "/var/lib/staticinator";
    cfg = config.services.staticinator;
  in
  {
    options.services.staticinator = {
      enable = mkEnableOption "Staticinator";

      package = mkPackageOption packages.${pkgs.stdenv.hostPlatform.system} "default" { };

      user = mkOption {
        type = str;
        default = "staticinator";
        description = "User account under which the bot runs.";
      };

      group = mkOption {
        type = str;
        default = "staticinator";
        description = "Group account under which the bot runs.";
      };

      port = mkOption {
        type = number;
        default = 7878;
        description = "The port that the application is hosted on.";
      };
    };

    config = mkIf cfg.enable {
      systemd.services = {
        staticinator = {
          description = "Staticinator";
          after = [ "network.target" ];
          wantedBy = [ "multi-user.target" ];
          restartTriggers = [
            cfg.package
            cfg.port
          ];
          environment = {
            DATA_PATH = dataPath;
            PORT = toString cfg.port;
          };

          serviceConfig = {
            Type = "simple";
            User = cfg.user;
            Group = cfg.group;
            StateDirectory = "staticinator";
            ExecStart = "${cfg.package}/bin/staticinator";
            Restart = "always";
          };
        };
      };

      users.users = optionalAttrs (cfg.user == "staticinator") {
        staticinator = {
          isSystemUser = true;
          group = cfg.group;
        };
      };

      users.groups = optionalAttrs (cfg.group == "staticinator") {
        staticinator = { };
      };

      environment.systemPackages = [
        (writeShellScriptBin "mkstatic" ''
          TOKEN=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 32)
          DIR="${dataPath}/$1"
          mkdir -m 0755 "$DIR"
          mkdir -m 0755 "$DIR/html"
          touch "$DIR/token"
          chmod 0600 "$DIR/token"
          echo "$TOKEN" >> "$DIR/token"
          chown -R ${cfg.user}:${cfg.group} "$DIR"
          echo "New static directory configured for $1, your token is:"
          echo "$TOKEN"
        '')
      ]
    };
  }
)