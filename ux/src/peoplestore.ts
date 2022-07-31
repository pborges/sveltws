import rpc from "./rpc"
import createStore, {Actionable, Keyed} from "./store"
import {nanoid} from "nanoid";

interface Name {
    first: string
    last: string
}

class Person extends Actionable<Person> implements Keyed {
    public firstName: string
    public lastName: string
    public age: number
    public children: Person[]

    public reset = () => this.action("person.reset", this.key())

    public key = (): Name => ({
        first: this.firstName,
        last: this.lastName,
    })
}

const people = createStore<Name, Person>(
    Person,
    "person.get",
    ({key, resolve, reject}) => {
        const subscriptionId: string = nanoid()
        rpc.send("person.get", {...key, subscribe: subscriptionId}).then(resolve).catch(reject)
        return () => rpc.send("unsubscribe", subscriptionId).then(resolve).catch(reject)
    }
)

export type {Person, Name}
export default people