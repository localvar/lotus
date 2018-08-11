var lotus = {
	getUrlArgs() {
		var res = new Object(), s = location.search;
		if (s.length == 0)
			return res;
		var fields = s.substring(1).split("&");
		for (var i = 0; i < fields.length; i++) {
			var kv = fields[i].split("=");
			res[kv[0]] = decodeURIComponent(kv[1]);
		}
		return res;
	},

	callApi(name, arg) {
		return new Promise((resolve, reject) => {
			axios.post('/api/'+name, arg).then(resp => {
				resp = resp.data
				if(resp.succeeded) {
					resolve(resp.data)
				} else {
					reject(new Error(resp.message))
				}
			}).catch(err=>{
				reject(err)
			})
		})
	}
}