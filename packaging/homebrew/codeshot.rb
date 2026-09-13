# Homebrew formula for codeshot.
#
# Used by the tap at securekomodo/homebrew-tap, and submitted to
# homebrew-core as Formula/c/codeshot.rb.
#
# After tagging a release, refresh url and sha256 with:
#
#   v=1.0.0
#   curl -fsSL -o /tmp/codeshot.tar.gz \
#     "https://github.com/securekomodo/codeshot/archive/refs/tags/v$v.tar.gz"
#   shasum -a 256 /tmp/codeshot.tar.gz
#
class Codeshot < Formula
  desc "Turn code into a beautiful image, from your terminal"
  homepage "https://github.com/securekomodo/codeshot"
  url "https://github.com/securekomodo/codeshot/archive/refs/tags/v0.0.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  # MIT for the program; the bundled fonts keep their own licenses.
  license all_of: ["MIT", "OFL-1.1", "Bitstream-Vera"]
  head "https://github.com/securekomodo/codeshot.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/codeshot --version")

    # Rendering is self-contained: no network, no browser, no system fonts.
    system bin/"codeshot", "--preset", "code", "--sample", "--output", testpath/"sample.png"
    assert_path_exists testpath/"sample.png"
    assert_equal "\x89PNG\r\n\x1a\n", (testpath/"sample.png").binread(8)

    system bin/"codeshot", "--preset", "git-diff", "--sample", "--output", testpath/"sample.svg"
    assert_match "<svg", (testpath/"sample.svg").read
  end
end
