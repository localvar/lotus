let lotus = {
    roles: [
        {value: 1, name: '黑名单用户'},
        {value: 2, name: '普通用户'},
        {value: 3, name: '编辑'},
        {value: 10, name: '管理员'},
    ],

    formatRole(role) {
        for(let r of lotus.roles)
            if(role == r.value)
                return r.name;
        return '';
    },

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