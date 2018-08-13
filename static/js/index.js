let lotus = {
	getUrlArgs() {
		let res = new Object(), s = location.search;
		if (s.length == 0)
			return res;
		let fields = s.substring(1).split("&");
		for (let i = 0; i < fields.length; i++) {
			let kv = fields[i].split("=");
			res[kv[0]] = decodeURIComponent(kv[1]);
		}
		return res;
	},

	callApi(name, arg) {
        return new Promise((resolve, reject) => {
            let url = '/api/' + name;

            // the next 4 lines is for debugging
            let uid = lotus.getUrlArgs().uid;
            if(uid) {
                url += '?uid=' + uid;
            }

			axios.post(url, arg).then(resp => {
				resp = resp.data;
				if(resp.succeeded) {
					resolve(resp.data);
				} else {
					reject(new Error(resp.message));
				}
			}).catch(err => {
				reject(err);
			});
		})
	},
};