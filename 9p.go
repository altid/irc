package main

import (
	"os"

	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/srv"
)

func setupFiles(st *state) (*srv.File, error) {
	user := p.OsUsers.Uid2User(os.Geteuid())
	root := new(srv.File)
	err := root.Add(nil, "/", user, nil, p.DMDIR|0777, nil)
	if err != nil {
		return nil, err
	}
	err = st.ctl.Add(root, "ctl", user, nil, 0666, st.ctl)
	if err != nil {
		return nil, err
	}
	err = st.current.Add(root, "main", user, nil, 0666, st.current)
	if err != nil {
		return nil, err
	}

	if st.input.show {
		err = st.input.Add(root, "input", user, nil, 0666, st.input)
		if err != nil {
			return nil, err
		}
	}
	if st.status.show {
		err = st.status.Add(root, "status", user, nil, 0644, st.status)
		if err != nil {
			return nil, err
		}
	}
	if st.title.show {
		err = st.title.Add(root, "title", user, nil, 0644, st.title)
		if err != nil {
			return nil, err
		}
	}
	if st.bar.show {
		err = st.bar.Add(root, "sidebar", user, nil, 0644, st.bar)
		if err != nil {
			return nil, err
		}
	}
	if st.tabs.show {
		err = st.tabs.Add(root, "tabs", user, nil, 0644, st.bar)
		if err != nil {
			return nil, err
		}
	}
	return root, nil

}
