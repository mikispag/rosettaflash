class X {

	static var app : X;

	function X(mc) {
		mc.loadMovie("http://miki.it/swf/rickroll.swf", 1);
	}

	// entry point
	static function main(mc) {
		app = new X(mc);
	}
}
	
