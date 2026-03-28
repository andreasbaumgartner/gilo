FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    curl \
    git \
    gcc \
    make \
    zsh \
    tmux \
    unzip \
    openssh-client \
    stow \
    ripgrep \
    bat \
    fd-find \
    xclip \
    && rm -rf /var/lib/apt/lists/*

# Symlink bat/fd (Ubuntu packages them with different names)
RUN ln -sf /usr/bin/batcat /usr/local/bin/bat \
    && ln -sf /usr/bin/fdfind /usr/local/bin/fd

# Install Go 1.24
RUN curl -fsSL https://go.dev/dl/go1.24.1.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

# Install neovim
RUN curl -fsSL https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.tar.gz | tar -C /usr/local --strip-components=1 -xz

# Install lazygit
RUN LAZYGIT_VERSION=$(curl -s "https://api.github.com/repos/jesseduffield/lazygit/releases/latest" | grep -Po '"tag_name": "v\K[^"]*') \
    && curl -fsSL "https://github.com/jesseduffield/lazygit/releases/latest/download/lazygit_${LAZYGIT_VERSION}_Linux_x86_64.tar.gz" | tar -C /usr/local/bin -xz lazygit

# Install fzf
RUN git clone --depth 1 https://github.com/junegunn/fzf.git /usr/local/share/fzf \
    && /usr/local/share/fzf/install --bin \
    && ln -sf /usr/local/share/fzf/bin/fzf /usr/local/bin/fzf

# Install zoxide
RUN curl -fsSL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | bash \
    && ln -sf /root/.local/bin/zoxide /usr/local/bin/zoxide 2>/dev/null; true

# Install Go development tools
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/fatih/gomodifytags@latest \
    && go install github.com/josharian/impl@latest \
    && go install github.com/cweill/gotests/gotests@latest \
    && go install mvdan.cc/gofumpt@latest

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin

# Set default shell to zsh
RUN chsh -s /bin/zsh root
ENV SHELL=/bin/zsh
