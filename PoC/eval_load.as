class X {

	static var app : X;

	function X(mc) {
        if (_root.eval) {
                eval(_root.eval);
        }
		if (_root.load) {
			mc.loadMovie(_root.load);
		}
	}

	// entry point
	static function main(mc) {
		app = new X(mc);
	}
}
	
