class Glitch < Formula
  desc "gl1tch workflow engine"
  homepage "https://8op.org"
  license "MIT"
  version "__VERSION__"

  depends_on "babashka"

  url "__BASE_URL__/glitch___VERSION__.tar.gz"
  sha256 "__SHA__"

  def install
    (share/"glitch").install "src"
    (share/"glitch/providers").install Dir["providers/*.clj"]
    (share/"glitch").install "ast-grep-rules"

    (bin/"glitch").write <<~SH
      #!/bin/bash
      exec bb -cp "#{share}/glitch/src" -m glitch.main "$@"
    SH
  end

  def post_install
    provider_dir = etc/"glitch/providers"
    provider_dir.mkpath
    (share/"glitch/providers").each_child do |f|
      provider_dir.install_symlink f unless (provider_dir/f.basename).exist?
    end
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/glitch version")
  end
end
