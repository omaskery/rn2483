package fake

func invalidParam(ctx *commandContext) error {
	return ctx.WriteResponse("invalid_param")
}

func ok(ctx *commandContext) error {
	return ctx.WriteResponse("ok")
}
