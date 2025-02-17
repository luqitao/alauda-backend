# gomod.alauda.cn/alauda-backend/pkg/registry

A common registration for `*restful.WebService` instances or constructor functions to decentralize initiation from wrapping up code.

Used to split different webservices implementation in different packages and/or files and use a centralized registry to connect them all on start up.