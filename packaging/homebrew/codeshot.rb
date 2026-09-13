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
  url "https://github.com/securekomodo/codeshot/archive/refs/tags/v1.0.1.tar.gz"
  sha256 "e32144149e7427e03c9706803127a561c62f4832bcbe24e30e45b846ad4154f6"
  # MIT for the program; the bundled fonts keep their own licenses.
  license all_of: ["MIT", "OFL-1.1", "Bitstream-Vera"]
  head "https://github.com/securekomodo/codeshot.git", branch: "main"

  depends_on "go" => :build

  def install
    # std_go_args already supplies -s -w and the output path.
    system "go", "build", *std_go_args(ldflags: "-X main.version=#{version}"), "./cmd/codeshot"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/codeshot --version")

    # Rendering is self-contained: no network, no browser, no system fonts.
    system bin/"codeshot", "--preset", "code", "--sample", "--output", testpath/"sample.png"
    assert_path_exists testpath/"sample.png"
    assert_equal "\x89PNG\r\n\x1a\n".b, (testpath/"sample.png").binread(8)

    system bin/"codeshot", "--preset", "git-diff", "--sample", "--output", testpath/"sample.svg"
    assert_match "<svg", (testpath/"sample.svg").read
  end
end
