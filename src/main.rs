use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "yuiop", about = "The last package manager you need to know", version)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Command,
}

#[derive(Subcommand)]
pub enum Command {
    /// Install a package
    Install { package: String },
    /// Remove a package
    Remove { package: String },
    /// List installed packages
    List,
    /// Search packages
    Search { term: String },
    /// Upgrade packages
    Upgrade {
        /// Upgrade only this package
        package: Option<String>,
        /// Upgrade all
        #[arg(long)]
        all: bool,
    },
    /// Package details
    Info { package: String },
    /// Show or set the platform override
    Platform {
        /// macos, debian or arch
        platform: Option<String>,
    },
    /// Show yuiop version
    Version,
}

fn main() {
    let cli = Cli::parse();
    match cli.command {
        Command::Install { package } => eprintln!("yuiop: install not implemented yet ({})", package),
        Command::Remove { package } => eprintln!("yuiop: remove not implemented yet ({})", package),
        Command::List => eprintln!("yuiop: list not implemented yet"),
        Command::Search { term } => eprintln!("yuiop: search not implemented yet ({})", term),
        Command::Upgrade { package, all } => eprintln!("yuiop: upgrade not implemented yet ({:?}, all={})", package, all),
        Command::Info { package } => eprintln!("yuiop: info not implemented yet ({})", package),
        Command::Platform { platform } => match platform {
            Some(p) => eprintln!("yuiop: platform override '{}' not persisted yet", p),
            None => println!("auto"),
        },
        Command::Version => println!("yuiop {}", env!("CARGO_PKG_VERSION")),
    }
}