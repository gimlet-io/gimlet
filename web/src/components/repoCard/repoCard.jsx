import { StarIcon } from "@heroicons/react/20/solid";
import { StarIcon as StarIconOutline } from "@heroicons/react/24/outline";
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';

function RepoCard(props) {
  const {name, services, navigateToRepo, favorite, favoriteHandler} = props;
  const numberOfServicesOnCard = 5;
  const truncatedServices = services.length > numberOfServicesOnCard ? services.slice(0, numberOfServicesOnCard) : services;

  const serviceWidgets = truncatedServices.map(service => {
    let ingressWidgets = [];
    if (service.ingresses !== undefined) {
      ingressWidgets = service.ingresses.map(ingress => (
        <p className="externalLink" key={service.service.name+"-"+ingress.url}>
          <span>{ingress.url}</span>
          <a href={`https://${ingress.url}`} target="_blank" rel="noopener noreferrer"
             onClick={(e) => {
               e.stopPropagation();
               return true
             }}>
            <ArrowTopRightOnSquareIcon className="externalLinkIcon ml-1" aria-hidden="true" />
          </a>
        </p>
      ))
    }

    return (
      <div key={service.service.name}>
        <p className="text-xs">{service.service.name}
          <span className="ml-1 badge">{service.env}</span>
        </p>
        <ul className="text-xs pl-2">
          {ingressWidgets}
        </ul>
      </div>
    )
  })

  return (
    <div className="relative">
      <div className="absolute top-0 right-0 m-6">
        {favorite &&
        <StarIcon
          className="h-4 w-4 text-teal-400 hover:text-teal-300 dark:hover:text-teal-500 cursor-pointer"
          onClick={() => favoriteHandler(name)}
        />
        }
        {!favorite &&
        <StarIconOutline
          className="h-4 w-4 text-neutral-400 hover:text-teal-400 cursor-pointer"
          onClick={() => favoriteHandler(name)}
        />
        }
      </div>
      <div className="w-full flex items-center justify-between p-6 space-x-6 cursor-pointer text-neutral-900 dark:text-neutral-200"
           onClick={() => navigateToRepo(name)}>
        <div className="flex-1 truncate">
          <p className="text-sm font-bold">{name}</p>
          <div className="p-2 space-y-2">
            {serviceWidgets}
            {services.length > numberOfServicesOnCard ?  <p>...</p> : null}
          </div>
        </div>
      </div>
    </div>
  )
}

export default RepoCard;
