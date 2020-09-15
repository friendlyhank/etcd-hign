package embed

func StartEtcd(){

	defer func() {
	}()

	//集群Listeners
	configurePeerListeners()

	//客户端Listeners
	configureClientListeners()
}

// configurePeerListeners - 设置集群的监听
func configurePeerListeners(){

}

// configureClientListeners -设置客户端的监听
func configureClientListeners(){

}
